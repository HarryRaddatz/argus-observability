package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func TestPurgeRemovesOldLogsAndMetrics(t *testing.T) {
	st, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	ctx := context.Background()
	now := time.Now().UTC()
	oldLog := now.Add(-10 * 24 * time.Hour)
	oldMetric := now.Add(-40 * 24 * time.Hour)

	if err := st.WriteLogs(ctx, []model.LogEntry{{
		TS: oldLog, Message: "old", Level: "info", EntityUID: "docker:h:c1",
	}}); err != nil {
		t.Fatal(err)
	}
	if err := st.WriteLogs(ctx, []model.LogEntry{{
		TS: now, Message: "new", Level: "info", EntityUID: "docker:h:c1",
	}}); err != nil {
		t.Fatal(err)
	}
	if err := st.WriteMetrics(ctx, []model.MetricPoint{{
		MetricName: "cpu.usage", TS: oldMetric, Value: 1, EntityUID: "docker:h:c1",
		Labels: model.Labels{"container": "c1"},
	}}); err != nil {
		t.Fatal(err)
	}
	if err := st.WriteMetrics(ctx, []model.MetricPoint{{
		MetricName: "cpu.usage", TS: now, Value: 2, EntityUID: "docker:h:c1",
		Labels: model.Labels{"container": "c1"},
	}}); err != nil {
		t.Fatal(err)
	}

	pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := st.Purge(pctx, now.Add(-7*24*time.Hour), now.Add(-30*24*time.Hour), now.Add(-30*24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if res.LogsDeleted != 1 {
		t.Fatalf("logs deleted: %d", res.LogsDeleted)
	}
	if res.MetricsDeleted != 1 {
		t.Fatalf("metrics deleted: %d", res.MetricsDeleted)
	}

	logs, err := st.SearchLogs(ctx, model.LogSearchFilter{Since: now.Add(-24 * time.Hour), Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || logs[0].Message != "new" {
		t.Fatalf("expected 1 recent log, got %+v", logs)
	}
}

func TestPurgeRespectsContextTimeout(t *testing.T) {
	st, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	ctx := context.Background()
	base := time.Now().UTC().Add(-365 * 24 * time.Hour)
	var entries []model.LogEntry
	for i := 0; i < 5000; i++ {
		entries = append(entries, model.LogEntry{
			TS: base.Add(time.Duration(i) * time.Second), Message: "x", Level: "info",
			EntityUID: "docker:h:c",
		})
	}
	if err := st.WriteLogs(ctx, entries); err != nil {
		t.Fatal(err)
	}

	pctx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()
	time.Sleep(2 * time.Millisecond)
	_, err = st.Purge(pctx, time.Now().UTC(), time.Now().UTC(), time.Now().UTC())
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
