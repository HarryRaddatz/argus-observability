package insights

import (
	"testing"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func TestDeriveMetricsFromLogExit(t *testing.T) {
	ts := time.Now().UTC()
	entry := model.LogEntry{
		TS:        ts,
		Message:   `{"event":"exit","service":"demo-api","status":200,"durationMs":250}`,
		EntityUID: "container:stack-demo-api-1",
		Labels:    model.Labels{"container": "stack-demo-api-1"},
	}
	points := DeriveMetricsFromLog(entry)
	if len(points) < 4 {
		t.Fatalf("expected at least 4 points, got %d", len(points))
	}
	names := map[string]float64{}
	for _, p := range points {
		names[p.MetricName] = p.Value
		if p.Labels["service"] != "demo-api" {
			t.Fatalf("service label missing: %+v", p.Labels)
		}
	}
	if names["http.duration_ms"] != 250 {
		t.Fatalf("duration: %v", names)
	}
	if names["http.error_rate"] != 0 {
		t.Fatalf("error rate should be 0: %v", names)
	}
}

func TestDeriveMetricsFromLogError(t *testing.T) {
	entry := model.LogEntry{
		TS:        time.Now().UTC(),
		Message:   `{"event":"exit","service":"checkout","status":503,"durationMs":50}`,
		EntityUID: "container:stack-demo-checkout-1",
		Labels:    model.Labels{"container": "stack-demo-checkout-1"},
	}
	points := DeriveMetricsFromLog(entry)
	var errRate float64
	for _, p := range points {
		if p.MetricName == "http.error_rate" {
			errRate = p.Value
		}
	}
	if errRate != 1 {
		t.Fatalf("expected error_rate=1, got %v", errRate)
	}
}
