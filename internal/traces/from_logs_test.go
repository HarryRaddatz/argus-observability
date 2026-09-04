package traces_test

import (
	"testing"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/HarryRaddatz/argus-observability/internal/traces"
)

func TestBuildFromLogsHTTPExit(t *testing.T) {
	ts := time.Now().UTC()
	logs := []model.LogEntry{{
		TS:        ts,
		Message:   `{"event":"exit","service":"agendamentoapi","method":"GET","url":"/api/x","status":200,"durationMs":120,"traceId":"50e30959-f59e-487f-bbe0-89ae7d8e74e5"}`,
		Level:     "info",
		EntityUID: "docker:host:venuz-agendamentoapi",
		Labels:    model.Labels{"container": "venuz-agendamentoapi", "service": "agendamentoapi"},
		Fields:    map[string]any{"trace_id": "50e30959f59e487fbbe089aed78e74e5"},
	}}
	detail := traces.BuildFromLogs("50e30959-f59e-487f-bbe0-89ae7d8e74e5", logs)
	if len(detail.Spans) != 1 {
		t.Fatalf("spans: %d", len(detail.Spans))
	}
	if detail.Spans[0].DurationMs != 120 {
		t.Fatalf("duration: %v", detail.Spans[0].DurationMs)
	}
	if detail.Source != "logs" {
		t.Fatalf("source: %s", detail.Source)
	}
}
