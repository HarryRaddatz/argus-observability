package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func TestListWorkloadsKeepsLatestPerContainer(t *testing.T) {
	st := openTest(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	point := func(name string, ts time.Time, v float64, container string) model.MetricPoint {
		return model.MetricPoint{
			MetricName: name, TS: ts, Value: v, EntityUID: "docker:host:" + container,
			Labels: model.Labels{"host": "host", "container": container, "service": "svc-" + container},
		}
	}
	if err := st.WriteMetrics(ctx, []model.MetricPoint{
		point("cpu.usage", now.Add(-2*time.Minute), 10, "api"),
		point("cpu.usage", now.Add(-time.Minute), 30, "api"),
		point("memory.usage", now.Add(-time.Minute), 512, "api"),
		point("memory.limit", now.Add(-time.Minute), 1024, "api"),
		point("cpu.usage", now.Add(-3*time.Hour), 99, "old"),
		point("http.requests", now, 5, "api"),
	}); err != nil {
		t.Fatal(err)
	}
	// An older batch arriving late must not overwrite the newer value.
	if err := st.WriteMetrics(ctx, []model.MetricPoint{point("cpu.usage", now.Add(-5*time.Minute), 1, "api")}); err != nil {
		t.Fatal(err)
	}

	got, err := st.ListWorkloads(ctx, now.Add(-15*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("want only the container seen in the window, got %+v", got)
	}
	w := got[0]
	if w.Container != "api" || w.CPUUsage != 30 || w.MemoryUsage != 512 || w.MemoryLimit != 1024 || w.Service != "svc-api" {
		t.Fatalf("unexpected snapshot: %+v", w)
	}

	wide, err := st.ListWorkloads(ctx, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(wide) != 2 {
		t.Fatalf("want both containers over 24h, got %+v", wide)
	}
}

func TestRecordLogPatternsUpsertsSample(t *testing.T) {
	st := openTest(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	entry := func(msg string, ts time.Time) model.LogEntry {
		return model.LogEntry{TS: ts, Message: msg, Level: "error", EntityUID: "docker:host:api",
			Labels: model.Labels{"container": "api", "service": "api"}}
	}
	if err := st.RecordLogPatterns(ctx, []model.LogEntry{entry("timeout after 120 ms", now.Add(-time.Minute))}); err != nil {
		t.Fatal(err)
	}
	if err := st.RecordLogPatterns(ctx, []model.LogEntry{entry("timeout after 340 ms", now)}); err != nil {
		t.Fatal(err)
	}
	got, err := st.ListLogPatterns(ctx, now.Add(-time.Hour), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Count != 2 || got[0].Sample != "timeout after 340 ms" {
		t.Fatalf("unexpected patterns: %+v", got)
	}
}
