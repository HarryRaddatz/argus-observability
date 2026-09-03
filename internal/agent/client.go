package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

type Config struct {
	HubURL      string
	AgentToken  string
	AgentID     string
	HostID      string
	Runtime     string
	Interval    time.Duration
	HTTPClient  *http.Client
}

type Client struct {
	cfg    Config
	logger *slog.Logger
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
	return &Client{cfg: cfg, logger: logger}
}

func (c *Client) Register(ctx context.Context, labels model.Labels) (*model.AgentSession, error) {
	body, _ := json.Marshal(map[string]any{
		"agent_id": c.cfg.AgentID,
		"host_id":  c.cfg.HostID,
		"runtime":  c.cfg.Runtime,
		"labels":   labels,
	})
	var session model.AgentSession
	if err := c.post(ctx, "/api/v1/agents/register", body, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (c *Client) Heartbeat(ctx context.Context) error {
	body, _ := json.Marshal(map[string]string{"agent_id": c.cfg.AgentID})
	return c.post(ctx, "/api/v1/agents/heartbeat", body, nil)
}

func (c *Client) SendMetrics(ctx context.Context, points []model.MetricPoint) error {
	if len(points) == 0 {
		return nil
	}
	body, err := json.Marshal(points)
	if err != nil {
		return err
	}
	return c.post(ctx, "/api/v1/metrics/batch", body, nil)
}

func (c *Client) SendLogs(ctx context.Context, entries []model.LogEntry) error {
	if len(entries) == 0 {
		return nil
	}
	body, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	return c.post(ctx, "/api/v1/logs/batch", body, nil)
}

func (c *Client) SendEvent(ctx context.Context, evt model.Event) error {
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return c.post(ctx, "/api/v1/events", body, nil)
}

func (c *Client) post(ctx context.Context, path string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.HubURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
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
		return fmt.Errorf("hub %s: %s", resp.Status, string(b))
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
