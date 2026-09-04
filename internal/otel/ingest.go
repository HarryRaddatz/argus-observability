package otel

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/google/uuid"
)

type otlpPayload struct {
	ResourceSpans []resourceSpans `json:"resourceSpans"`
}

type resourceSpans struct {
	Resource   otlpResource `json:"resource"`
	ScopeSpans []scopeSpans `json:"scopeSpans"`
}

type otlpResource struct {
	Attributes []otlpAttr `json:"attributes"`
}

type scopeSpans struct {
	Spans []otlpSpan `json:"spans"`
}

type otlpSpan struct {
	TraceID           string     `json:"traceId"`
	SpanID            string     `json:"spanId"`
	ParentSpanID      string     `json:"parentSpanId"`
	Name              string     `json:"name"`
	Kind              int        `json:"kind"`
	StartTimeUnixNano string     `json:"startTimeUnixNano"`
	EndTimeUnixNano   string     `json:"endTimeUnixNano"`
	Attributes        []otlpAttr `json:"attributes"`
	Status            struct {
		Code int `json:"code"`
	} `json:"status"`
}

type otlpAttr struct {
	Key   string `json:"key"`
	Value struct {
		StringValue string `json:"stringValue"`
		IntValue    string `json:"intValue"`
	} `json:"value"`
}

// ParseOTLPTracesJSON decodes OTLP/HTTP JSON trace export into Argus spans.
func ParseOTLPTracesJSON(body []byte) ([]model.TraceSpan, error) {
	var payload otlpPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	var out []model.TraceSpan
	for _, rs := range payload.ResourceSpans {
		service := attrString(rs.Resource.Attributes, "service.name")
		container := attrString(rs.Resource.Attributes, "container.id", "container.name")
		for _, ss := range rs.ScopeSpans {
			for _, sp := range ss.Spans {
				traceID := decodeID(sp.TraceID)
				spanID := decodeID(sp.SpanID)
				if traceID == "" || spanID == "" {
					continue
				}
				start := nanoTime(sp.StartTimeUnixNano)
				end := nanoTime(sp.EndTimeUnixNano)
				durMs := 0.0
				if !start.IsZero() && !end.IsZero() && end.After(start) {
					durMs = float64(end.Sub(start).Milliseconds())
				}
				attrs := map[string]any{}
				for _, a := range sp.Attributes {
					if v := attrVal(a); v != "" {
						attrs[a.Key] = v
					}
				}
				status := "unset"
				switch sp.Status.Code {
				case 1:
					status = "ok"
				case 2:
					status = "error"
				}
				out = append(out, model.TraceSpan{
					TraceID:      traceID,
					SpanID:       spanID,
					ParentSpanID: decodeID(sp.ParentSpanID),
					Name:         sp.Name,
					Service:      firstNonEmpty(service, attrString(sp.Attributes, "service.name")),
					Container:    container,
					StartTS:      start,
					EndTS:        end,
					DurationMs:   durMs,
					Status:       status,
					Kind:         spanKindName(sp.Kind),
					Source:       "otlp",
					Attributes:   attrs,
				})
			}
		}
	}
	return out, nil
}

func decodeID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if len(raw) == 32 && isHex(raw) {
		return strings.ToLower(raw)
	}
	if len(raw) == 36 && strings.Count(raw, "-") == 4 {
		return strings.ReplaceAll(strings.ToLower(raw), "-", "")
	}
	b, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		b, err = base64.RawStdEncoding.DecodeString(raw)
	}
	if err == nil && len(b) > 0 {
		return hex.EncodeToString(b)
	}
	return strings.ToLower(raw)
}

func isHex(s string) bool {
	for _, ch := range s {
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') && (ch < 'A' || ch > 'F') {
			return false
		}
	}
	return true
}

func nanoTime(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}
	}
	var n int64
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return time.Time{}
		}
		n = n*10 + int64(ch-'0')
	}
	if n <= 0 {
		return time.Time{}
	}
	return time.Unix(0, n).UTC()
}

func attrString(attrs []otlpAttr, keys ...string) string {
	for _, k := range keys {
		for _, a := range attrs {
			if a.Key == k {
				if v := attrVal(a); v != "" {
					return v
				}
			}
		}
	}
	return ""
}

func attrVal(a otlpAttr) string {
	if a.Value.StringValue != "" {
		return a.Value.StringValue
	}
	return a.Value.IntValue
}

func spanKindName(kind int) string {
	switch kind {
	case 1:
		return "internal"
	case 2:
		return "server"
	case 3:
		return "client"
	case 4:
		return "producer"
	case 5:
		return "consumer"
	default:
		return "unspecified"
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// NewLogSpanID generates a stable span id for log-derived spans.
func NewLogSpanID(container string, ts time.Time) string {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(container+"|"+ts.UTC().Format(time.RFC3339Nano))).String()[:16]
}
