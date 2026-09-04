package traces

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/insights"
	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/HarryRaddatz/argus-observability/internal/otel"
)

// BuildFromLogs synthesizes a waterfall from correlated log lines when OTLP spans are absent.
func BuildFromLogs(traceID string, logs []model.LogEntry) model.TraceDetail {
	spans := make([]model.TraceSpan, 0, len(logs))
	for _, log := range logs {
		tid := traceIDFromEntry(log)
		if tid == "" {
			continue
		}
		if traceID != "" && !traceIDsMatch(tid, traceID) {
			continue
		}
		container := log.Labels["container"]
		if container == "" {
			parts := strings.Split(log.EntityUID, ":")
			if len(parts) > 0 {
				container = parts[len(parts)-1]
			}
		}
		service := log.Labels["service"]
		if service == "" {
			service = insights.InferServiceFromContainer(container)
		}

		if sig, ok := insights.ParseHTTPLog(log.Message); ok && sig.IsExit {
			if sig.Service != "" {
				service = sig.Service
			}
			start := log.TS
			if sig.DurationMs > 0 {
				start = log.TS.Add(-time.Duration(sig.DurationMs) * time.Millisecond)
			}
			name := strings.TrimSpace(fmt.Sprintf("%s %s", sig.Method, sig.Path))
			if name == "" {
				name = "http.request"
			}
			status := "ok"
			if sig.Status >= 400 {
				status = "error"
			}
			spanID := spanIDFromFields(log)
			if spanID == "" {
				spanID = otel.NewLogSpanID(container, log.TS)
			}
			spans = append(spans, model.TraceSpan{
				TraceID:    normalizeTrace(traceID, tid),
				SpanID:     spanID,
				Name:       name,
				Service:    service,
				Container:  container,
				EntityUID:  log.EntityUID,
				StartTS:    start,
				EndTS:      log.TS,
				DurationMs: sig.DurationMs,
				Status:     status,
				Kind:       "server",
				Source:     "logs",
				Attributes: map[string]any{
					"http.status": sig.Status,
					"level":       log.Level,
				},
			})
			continue
		}

		name := shortMessage(log.Message)
		spanID := spanIDFromFields(log)
		if spanID == "" {
			spanID = otel.NewLogSpanID(container+"|log", log.TS)
		}
		spans = append(spans, model.TraceSpan{
			TraceID:    normalizeTrace(traceID, tid),
			SpanID:     spanID,
			Name:       name,
			Service:    service,
			Container:  container,
			EntityUID:  log.EntityUID,
			StartTS:    log.TS,
			EndTS:      log.TS,
			DurationMs: 0,
			Status:     logLevelStatus(log.Level),
			Kind:       "log",
			Source:     "logs",
			Attributes: map[string]any{"level": log.Level},
		})
	}

	sort.Slice(spans, func(i, j int) bool {
		if spans[i].StartTS.Equal(spans[j].StartTS) {
			return spans[i].EndTS.Before(spans[j].EndTS)
		}
		return spans[i].StartTS.Before(spans[j].StartTS)
	})

	detail := model.TraceDetail{
		TraceID: normalizeTrace(traceID, traceIDFromLogs(logs)),
		Source:  "logs",
		Spans:   spans,
	}
	if len(spans) > 0 {
		detail.StartTS = spans[0].StartTS
		detail.EndTS = spans[len(spans)-1].EndTS
		if detail.EndTS.After(detail.StartTS) {
			detail.DurationMs = float64(detail.EndTS.Sub(detail.StartTS).Milliseconds())
		}
	}
	return detail
}

func traceIDFromEntry(log model.LogEntry) string {
	if log.Fields != nil {
		for _, k := range []string{"trace_id", "traceId"} {
			if v, ok := log.Fields[k].(string); ok && v != "" {
				return normalizeTrace("", v)
			}
		}
	}
	for k, v := range insights.ParseTraceFields(log.Message) {
		if k == "trace_id" && v != "" {
			return normalizeTrace("", v)
		}
	}
	return ""
}

func spanIDFromFields(log model.LogEntry) string {
	if log.Fields == nil {
		return ""
	}
	for _, k := range []string{"span_id", "spanId"} {
		if v, ok := log.Fields[k].(string); ok {
			return strings.ToLower(strings.TrimSpace(v))
		}
	}
	return ""
}

func traceIDFromLogs(logs []model.LogEntry) string {
	for _, log := range logs {
		if tid := traceIDFromEntry(log); tid != "" {
			return tid
		}
	}
	return ""
}

func normalizeTrace(fallback, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	raw = strings.ToLower(raw)
	raw = strings.ReplaceAll(raw, "-", "")
	if len(raw) == 32 {
		return raw
	}
	return raw
}

func traceIDsMatch(a, b string) bool {
	return normalizeTrace(a, a) == normalizeTrace(b, b) ||
		strings.EqualFold(strings.ReplaceAll(a, "-", ""), strings.ReplaceAll(b, "-", ""))
}

func shortMessage(msg string) string {
	msg = strings.TrimSpace(msg)
	if strings.HasPrefix(msg, "{") {
		var m map[string]any
		if err := json.Unmarshal([]byte(msg), &m); err == nil {
			if event, _ := m["event"].(string); event != "" {
				if svc, _ := m["service"].(string); svc != "" {
					return svc + " " + event
				}
				return event
			}
		}
	}
	if len(msg) > 80 {
		return msg[:77] + "..."
	}
	return msg
}

func logLevelStatus(level string) string {
	switch strings.ToLower(level) {
	case "error", "fatal", "critical":
		return "error"
	case "warn", "warning":
		return "warn"
	default:
		return "ok"
	}
}
