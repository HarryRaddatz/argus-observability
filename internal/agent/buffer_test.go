package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func TestSpoolSurvivesRestartAndSkipsCorrupt(t *testing.T) {
	dir := t.TempDir()
	spool, err := OpenSpool(dir, 1<<20, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := spool.Put("/api/v1/metrics/batch", "b1", []byte(`[{"metric_name":"cpu.usage"}]`)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "0000000000000002.json"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := spool.Put("/api/v1/logs/batch", "b2", []byte(`[{"message":"x"}]`)); err != nil {
		t.Fatal(err)
	}

	again, err := OpenSpool(dir, 1<<20, nil)
	if err != nil {
		t.Fatal(err)
	}
	first, err := again.Front()
	if err != nil || first == nil || first.Path != "/api/v1/metrics/batch" {
		t.Fatalf("first = %+v err %v", first, err)
	}
	if err := again.Remove(first.name); err != nil {
		t.Fatal(err)
	}
	second, err := again.Front()
	if err != nil || second == nil || second.Path != "/api/v1/logs/batch" {
		t.Fatalf("second = %+v err %v", second, err)
	}
	bad, err := filepath.Glob(filepath.Join(dir, "*.bad"))
	if err != nil || len(bad) != 1 {
		t.Fatalf("corrupt files = %v err %v", bad, err)
	}
}

func TestSpoolDropsOldestWhenFull(t *testing.T) {
	dir := t.TempDir()
	spool, err := OpenSpool(dir, 1<<20, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := spool.Put("/api/v1/events", "a", []byte(`{"type":"a"}`)); err != nil {
		t.Fatal(err)
	}
	names, err := spool.names()
	if err != nil || len(names) != 1 {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(dir, names[0]))
	if err != nil {
		t.Fatal(err)
	}
	tight, err := OpenSpool(dir, info.Size()+8, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := tight.Put("/api/v1/events", "b", []byte(`{"type":"b"}`)); err != nil {
		t.Fatal(err)
	}
	item, err := tight.Front()
	if err != nil || item == nil {
		t.Fatal(err)
	}
	if string(item.Body) != `{"type":"b"}` {
		t.Fatalf("kept %s", item.Body)
	}
}

func TestSendBatchPersistsTransientAndDropsPermanent(t *testing.T) {
	dir := t.TempDir()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := bufferedClient(t, srv.URL, dir)
	err := c.SendMetrics(context.Background(), []model.MetricPoint{{MetricName: "cpu.usage"}})
	if err == nil {
		t.Fatal("expected error")
	}
	item, ferr := c.spool.Front()
	if ferr != nil || item == nil || item.Path != "/api/v1/metrics/batch" {
		t.Fatalf("buffered = %+v err %v", item, ferr)
	}

	hits.Store(0)
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		http.Error(w, "no", http.StatusUnauthorized)
	})
	c.drain(context.Background())
	if hits.Load() != 1 {
		t.Fatalf("replay hits = %d", hits.Load())
	}
	left, err := c.spool.Front()
	if err != nil {
		t.Fatal(err)
	}
	if left != nil {
		t.Fatalf("permanent failure kept %+v", left)
	}
}

func TestSendBatchDrainsBeforeNewBatch(t *testing.T) {
	dir := t.TempDir()
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	c := bufferedClient(t, srv.URL, dir)
	if err := c.spool.Put("/api/v1/logs/batch", "old-batch", []byte(`[{"message":"old"}]`)); err != nil {
		t.Fatal(err)
	}
	if err := c.SendMetrics(context.Background(), []model.MetricPoint{{MetricName: "cpu.usage"}}); err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 || paths[0] != "/api/v1/logs/batch" || paths[1] != "/api/v1/metrics/batch" {
		t.Fatalf("paths = %v", paths)
	}
	left, err := c.spool.Front()
	if err != nil || left != nil {
		t.Fatalf("spool = %+v err %v", left, err)
	}
}

func TestPermanentFailureIsNotBuffered(t *testing.T) {
	dir := t.TempDir()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "no", http.StatusForbidden)
	}))
	defer srv.Close()
	c := bufferedClient(t, srv.URL, dir)
	if err := c.SendLogs(context.Background(), []model.LogEntry{{Message: "x"}}); err == nil {
		t.Fatal("expected error")
	}
	item, err := c.spool.Front()
	if err != nil || item != nil {
		t.Fatalf("spool = %+v err %v", item, err)
	}
}

func TestReplayKeepsTheSameBatchID(t *testing.T) {
	dir := t.TempDir()
	var ids []string
	var mode atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ids = append(ids, r.Header.Get("X-Argus-Batch-Id"))
		if mode.Load() == 0 {
			http.Error(w, "down", http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()
	c := bufferedClient(t, srv.URL, dir)
	if err := c.SendMetrics(context.Background(), []model.MetricPoint{{MetricName: "cpu.usage", Value: 1}}); err == nil {
		t.Fatal("expected failure")
	}
	if len(ids) != retryAttempts {
		t.Fatalf("attempts = %d", len(ids))
	}
	for _, id := range ids[1:] {
		if id == "" || id != ids[0] {
			t.Fatalf("retry ids = %v", ids)
		}
	}
	stored, err := c.spool.Front()
	if err != nil || stored == nil || stored.BatchID != ids[0] {
		t.Fatalf("stored = %+v err %v", stored, err)
	}
	mode.Store(1)
	ids = nil
	if err := c.SendMetrics(context.Background(), []model.MetricPoint{{MetricName: "cpu.usage", Value: 2}}); err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] != stored.BatchID || ids[1] == ids[0] || ids[1] == "" {
		t.Fatalf("replay ids = %v", ids)
	}
}

func bufferedClient(t *testing.T, hub, dir string) *Client {
	t.Helper()
	c := NewClient(Config{
		HubURL:         hub,
		BufferDir:      dir,
		BufferMaxBytes: 1 << 20,
		HTTPClient:     &http.Client{Timeout: time.Second},
	}, nil)
	c.retryWait = func(context.Context, time.Duration) error { return nil }
	return c
}
