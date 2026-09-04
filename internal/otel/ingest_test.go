package otel_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/otel"
)

func TestParseOTLPTracesJSON(t *testing.T) {
	payload := map[string]any{
		"resourceSpans": []any{
			map[string]any{
				"resource": map[string]any{
					"attributes": []any{
						map[string]any{"key": "service.name", "value": map[string]any{"stringValue": "agendamentoapi"}},
					},
				},
				"scopeSpans": []any{
					map[string]any{
						"spans": []any{
							map[string]any{
								"traceId":           "50e30959f59e487fbbe089aed78e74e5",
								"spanId":            "a1b2c3d4e5f6a7b8",
								"name":              "GET /api/agenda",
								"kind":              2,
								"startTimeUnixNano": "1000000000",
								"endTimeUnixNano":   "2500000000",
								"status":            map[string]any{"code": 1},
							},
						},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(payload)
	spans, err := otel.ParseOTLPTracesJSON(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Service != "agendamentoapi" {
		t.Fatalf("service: %s", spans[0].Service)
	}
	if spans[0].DurationMs != 1 {
		t.Fatalf("duration: %v", spans[0].DurationMs)
	}
	if spans[0].StartTS != time.Unix(0, 1000000000).UTC() {
		t.Fatalf("start: %v", spans[0].StartTS)
	}
}
