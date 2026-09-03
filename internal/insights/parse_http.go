package insights

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

var venuzResponseRe = regexp.MustCompile(`(?i)(get|post|put|delete|patch|head|options)\s+(\S+)\s+.*?\s(\d{3})\s+\((\d+)ms\)\s*$`)

// HTTPLogSignal holds structured HTTP data extracted from a Venuz log line.
type HTTPLogSignal struct {
	Service    string
	Method     string
	Path       string
	Status     int
	DurationMs float64
	IsExit     bool
}

// ParseHTTPLog extracts HTTP request/response signals from JSON or Venuz response lines.
func ParseHTTPLog(message string) (HTTPLogSignal, bool) {
	msg := strings.TrimSpace(message)
	if msg == "" {
		return HTTPLogSignal{}, false
	}

	if strings.HasPrefix(msg, "{") {
		var m map[string]any
		if err := json.Unmarshal([]byte(msg), &m); err == nil {
			event, _ := m["event"].(string)
			if !strings.EqualFold(event, "exit") {
				return HTTPLogSignal{}, false
			}
			sig := HTTPLogSignal{
				IsExit:     true,
				Service:    stringField(m, "service"),
				Method:     strings.ToUpper(stringField(m, "method")),
				Path:       stringField(m, "url", "path"),
				Status:     intField(m, "status", "statusCode", "status_code"),
				DurationMs: floatField(m, "durationMs", "duration_ms", "duration"),
			}
			return sig, sig.Status > 0 || sig.DurationMs > 0
		}
	}

	if m := venuzResponseRe.FindStringSubmatch(msg); len(m) >= 5 {
		status, _ := strconv.Atoi(m[3])
		dur, _ := strconv.ParseFloat(m[4], 64)
		return HTTPLogSignal{
			Method:     strings.ToUpper(m[1]),
			Path:       m[2],
			Status:     status,
			DurationMs: dur,
			IsExit:     true,
		}, true
	}

	return HTTPLogSignal{}, false
}

func stringField(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case string:
				if s := strings.TrimSpace(t); s != "" {
					return s
				}
			}
		}
	}
	return ""
}

func intField(m map[string]any, keys ...string) int {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case float64:
				return int(t)
			case int:
				return t
			case json.Number:
				if i, err := t.Int64(); err == nil {
					return int(i)
				}
			}
		}
	}
	return 0
}

func floatField(m map[string]any, keys ...string) float64 {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case float64:
				return t
			case int:
				return float64(t)
			case json.Number:
				if f, err := t.Float64(); err == nil {
					return f
				}
			}
		}
	}
	return 0
}

func inferServiceFromContainer(container string) string {
	c := strings.TrimPrefix(container, "venuz-")
	if idx := strings.LastIndex(c, "-"); idx > 0 {
		suffix := c[idx+1:]
		if _, err := strconv.Atoi(suffix); err == nil {
			return c[:idx]
		}
	}
	return c
}
