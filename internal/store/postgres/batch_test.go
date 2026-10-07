package postgres

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func TestApplyBatchDuplicateAndConcurrent(t *testing.T) {
	st := openTest(t)
	now := time.Now().UTC().Truncate(time.Second)
	points := []model.MetricPoint{{
		MetricName: "cpu.usage",
		TS:         now,
		Value:      1,
		EntityUID:  "docker:host:api",
		Labels:     model.Labels{"container": "api"},
	}}
	write := func(ctx context.Context) error {
		return st.WriteMetrics(ctx, points)
	}
	applied, err := st.ApplyBatch(context.Background(), "batch-1", write)
	if err != nil || !applied {
		t.Fatalf("first applied=%v err=%v", applied, err)
	}
	applied, err = st.ApplyBatch(context.Background(), "batch-1", write)
	if err != nil || applied {
		t.Fatalf("replay applied=%v err=%v", applied, err)
	}
	got, err := st.QueryMetrics(context.Background(), "cpu.usage", nil, now.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("points = %d", len(got))
	}

	var wg sync.WaitGroup
	start := make(chan struct{})
	var wins atomic.Int32
	other := []model.MetricPoint{{
		MetricName: "cpu.usage",
		TS:         now.Add(time.Second),
		Value:      2,
		EntityUID:  "docker:host:api",
		Labels:     model.Labels{"container": "api"},
	}}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			ok, err := st.ApplyBatch(context.Background(), "batch-2", func(ctx context.Context) error {
				return st.WriteMetrics(ctx, other)
			})
			if err != nil {
				t.Errorf("concurrent: %v", err)
				return
			}
			if ok {
				wins.Add(1)
			}
		}()
	}
	close(start)
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatalf("wins = %d", wins.Load())
	}
	got, err = st.QueryMetrics(context.Background(), "cpu.usage", nil, now.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("points after concurrent = %d", len(got))
	}
}
