package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func TestWriteAndQueryMetrics(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	now := time.Now().UTC().Truncate(time.Second)
	points := []model.MetricPoint{{
		MetricName: "cpu.usage",
		TS:         now,
		Value:      55.5,
		EntityUID:  "docker:host:api",
		Labels:     model.Labels{"host": "host", "container": "api"},
	}}
	if err := st.WriteMetrics(context.Background(), points); err != nil {
		t.Fatal(err)
	}
	got, err := st.QueryMetrics(context.Background(), "cpu.usage", nil, now.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Value != 55.5 {
		t.Fatalf("unexpected series: %+v", got)
	}
}
