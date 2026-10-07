package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/google/uuid"
)

type Config struct {
	HubURL         string
	AgentToken     string
	AgentID        string
	HostID         string
	Runtime        string
	Interval       time.Duration
	HTTPClient     *http.Client
	BufferDir      string
	BufferMaxBytes int64
	ReplayTimeout  time.Duration
}

type Client struct {
	cfg       Config
	logger    *slog.Logger
	retryWait func(context.Context, time.Duration) error
	spool     *Spool
	batchMu   sync.Mutex
}

func NewClient(cfg Config, logger *slog.Logger) *Client {
	if logger == nil {
		logger = slog.Default()
	}
	if cfg.Interval == 0 {
		cfg.Interval = 15 * time.Second
	}
	if cfg.Runtime == "" {
		cfg.Runtime = "docker"
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 15 * time.Second}
	}
	c := &Client{cfg: cfg, logger: logger}
	if cfg.BufferDir != "" {
		maxBytes := cfg.BufferMaxBytes
		if maxBytes <= 0 {
			maxBytes = 64 << 20
		}
		spool, err := OpenSpool(cfg.BufferDir, maxBytes, logger)
		if err != nil {
			logger.Error("open buffer", "dir", cfg.BufferDir, "err", err)
		} else {
			c.spool = spool
		}
	}
	return c
}

func (c *Client) Register(ctx context.Context, labels model.Labels) (*model.AgentSession, error) {
	body, _ := json.Marshal(map[string]any{
		"agent_id": c.cfg.AgentID,
		"host_id":  c.cfg.HostID,
		"runtime":  c.cfg.Runtime,
		"labels":   labels,
	})
	var session model.AgentSession
	if err := c.post(ctx, "/api/v1/agents/register", "", body, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (c *Client) Heartbeat(ctx context.Context) error {
	body, _ := json.Marshal(map[string]string{"agent_id": c.cfg.AgentID})
	return c.post(ctx, "/api/v1/agents/heartbeat", "", body, nil)
}

func (c *Client) SendMetrics(ctx context.Context, points []model.MetricPoint) error {
	if len(points) == 0 {
		return nil
	}
	body, err := json.Marshal(points)
	if err != nil {
		return err
	}
	return c.sendBatch(ctx, "/api/v1/metrics/batch", uuid.NewString(), body)
}

func (c *Client) SendLogs(ctx context.Context, entries []model.LogEntry) error {
	if len(entries) == 0 {
		return nil
	}
	body, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	return c.sendBatch(ctx, "/api/v1/logs/batch", uuid.NewString(), body)
}

func (c *Client) SendEvent(ctx context.Context, evt model.Event) error {
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return c.sendBatch(ctx, "/api/v1/events", uuid.NewString(), body)
}

func (c *Client) SendFleet(ctx context.Context, rows []model.ContainerFleetStatus) error {
	if len(rows) == 0 {
		return nil
	}
	body, err := json.Marshal(rows)
	if err != nil {
		return err
	}
	return c.sendBatch(ctx, "/api/v1/fleet/batch", uuid.NewString(), body)
}

func (c *Client) SendTopology(ctx context.Context, links []model.TopologyLink) error {
	if len(links) == 0 {
		return nil
	}
	body, err := json.Marshal(links)
	if err != nil {
		return err
	}
	return c.sendBatch(ctx, "/api/v1/topology/batch", uuid.NewString(), body)
}

func (c *Client) post(ctx context.Context, path, batchID string, body []byte, out any) error {
	var last error
	for attempt := 1; attempt <= retryAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			if last != nil {
				return last
			}
			return err
		}
		if attempt > 1 {
			wait := c.retryWait
			if wait == nil {
				wait = waitContext
			}
			if err := wait(ctx, backoffDelay(attempt-1, rand.Int64N)); err != nil {
				return last
			}
		}
		last = c.postOnce(ctx, path, batchID, body, out)
		if last == nil || !retryable(last) {
			return last
		}
	}
	return last
}

// sendBatch drains older batches first, then posts this one through the same retrying transport.
// batchID is fixed for the life of the payload, including a later replay.
// A transient failure is stored. A permanent 4xx is not.
func (c *Client) sendBatch(ctx context.Context, path, batchID string, body []byte) error {
	if c.spool != nil {
		c.batchMu.Lock()
		defer c.batchMu.Unlock()
		c.drain(ctx)
	}
	err := c.post(ctx, path, batchID, body, nil)
	if err == nil || c.spool == nil || !retryable(err) {
		return err
	}
	if putErr := c.spool.Put(path, batchID, body); putErr != nil {
		c.logger.Error("buffer batch", "path", path, "err", putErr)
		return err
	}
	c.logger.Warn("buffered batch", "path", path, "batch_id", batchID, "bytes", len(body))
	return err
}

// drain replays the spool in order. Each item uses its own timeout.
// A transient failure keeps the item and stops the drain. A permanent failure drops it.
func (c *Client) drain(parent context.Context) {
	if c.spool == nil || parent.Err() != nil {
		return
	}
	for {
		item, err := c.spool.Front()
		if err != nil {
			c.logger.Warn("read buffer", "err", err)
			return
		}
		if item == nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), c.replayTimeout())
		err = c.post(ctx, item.Path, item.BatchID, item.Body, nil)
		cancel()
		if err == nil {
			if rmErr := c.spool.Remove(item.name); rmErr != nil {
				c.logger.Error("remove buffered batch", "file", item.name, "err", rmErr)
				return
			}
			continue
		}
		if !retryable(err) {
			c.logger.Warn("drop buffered batch", "path", item.Path, "err", err)
			if rmErr := c.spool.Remove(item.name); rmErr != nil {
				c.logger.Error("remove buffered batch", "file", item.name, "err", rmErr)
			}
			continue
		}
		return
	}
}

func (c *Client) replayTimeout() time.Duration {
	if c.cfg.ReplayTimeout > 0 {
		return c.cfg.ReplayTimeout
	}
	if c.cfg.HTTPClient != nil && c.cfg.HTTPClient.Timeout > 0 {
		return c.cfg.HTTPClient.Timeout
	}
	return 15 * time.Second
}

func (c *Client) postOnce(ctx context.Context, path, batchID string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.HubURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if batchID != "" {
		req.Header.Set("X-Argus-Batch-Id", batchID)
	}
	if c.cfg.AgentToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.AgentToken)
	}
	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return &hubStatusError{code: resp.StatusCode, status: resp.Status, detail: string(b)}
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *Client) Interval() time.Duration {
	return c.cfg.Interval
}

func (c *Client) AgentID() string { return c.cfg.AgentID }
func (c *Client) HostID() string  { return c.cfg.HostID }
