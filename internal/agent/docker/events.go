package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/google/uuid"
)

type dockerEvent struct {
	Type   string `json:"Type"`
	Action string `json:"Action"`
	Actor  struct {
		ID         string            `json:"ID"`
		Attributes map[string]string `json:"Attributes"`
	} `json:"Actor"`
	Time     int64 `json:"time"`
	TimeNano int64 `json:"timeNano"`
}

// StreamEvents follows the Docker events API and maps container actions to Argus events.
func (c *Collector) StreamEvents(ctx context.Context, emit func(model.Event) error) error {
	since := time.Now().UTC().Unix()
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := c.followEvents(ctx, since, emit)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
			}
		}
		since = time.Now().UTC().Unix()
	}
}

func (c *Collector) followEvents(ctx context.Context, since int64, emit func(model.Event) error) error {
	url := fmt.Sprintf("http://docker/v1.44/events?since=%d&filters=%s", since, eventFilters())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.streamClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return fmt.Errorf("docker events: %s %s", resp.Status, b)
	}

	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 64*1024), 256*1024)
	for sc.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var raw dockerEvent
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}
		if raw.Type != "container" {
			continue
		}
		name := containerNameFromEvent(raw)
		if name == "" {
			continue
		}
		if prefix := os.Getenv("ARGUS_NAME_PREFIX"); prefix != "" && !strings.HasPrefix(name, prefix) {
			continue
		}
		evt, ok := mapDockerEvent(c.hostID, name, raw)
		if !ok {
			continue
		}
		if err := emit(evt); err != nil {
			return err
		}
		if raw.Time > since {
			since = raw.Time
		}
	}
	return sc.Err()
}

func eventFilters() string {
	return `{"type":["container"]}`
}

func containerNameFromEvent(raw dockerEvent) string {
	if n := raw.Actor.Attributes["name"]; n != "" {
		return strings.TrimPrefix(n, "/")
	}
	return ""
}

func mapDockerEvent(hostID, name string, raw dockerEvent) (model.Event, bool) {
	entityUID := fmt.Sprintf("docker:%s:%s", hostID, name)
	labels := eventLabels(hostID, name, raw.Actor.Attributes)

	ts := time.Unix(raw.Time, 0).UTC()
	if raw.TimeNano > 0 {
		ts = time.Unix(0, raw.TimeNano).UTC()
	}

	payload := map[string]any{
		"action":       raw.Action,
		"container_id": raw.Actor.ID,
	}
	for k, v := range raw.Actor.Attributes {
		if k == "exitCode" || k == "signal" || k == "image" {
			payload[k] = v
		}
	}

	switch raw.Action {
	case "start":
		return model.Event{
			ID: uuid.NewString(), Type: "container.start", TS: ts,
			Severity: "info", Source: "agent", EntityUID: entityUID, Labels: labels, Payload: payload,
		}, true
	case "die":
		sev := "warning"
		if raw.Actor.Attributes["exitCode"] == "0" {
			sev = "info"
		}
		return model.Event{
			ID: uuid.NewString(), Type: "container.die", TS: ts,
			Severity: sev, Source: "agent", EntityUID: entityUID, Labels: labels, Payload: payload,
		}, true
	case "oom":
		return model.Event{
			ID: uuid.NewString(), Type: "container.oom", TS: ts,
			Severity: "critical", Source: "agent", EntityUID: entityUID, Labels: labels, Payload: payload,
		}, true
	case "restart":
		return model.Event{
			ID: uuid.NewString(), Type: "container.restart", TS: ts,
			Severity: "warning", Source: "agent", EntityUID: entityUID, Labels: labels, Payload: payload,
		}, true
	case "pause", "unpause", "destroy", "rename":
		return model.Event{
			ID: uuid.NewString(), Type: "container." + raw.Action, TS: ts,
			Severity: "info", Source: "agent", EntityUID: entityUID, Labels: labels, Payload: payload,
		}, true
	default:
		return model.Event{}, false
	}
}
