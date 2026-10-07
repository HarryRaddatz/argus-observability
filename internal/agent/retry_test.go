package agent

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func TestBackoffDelayGrowsThenCaps(t *testing.T) {
	noJitter := func(n int64) int64 { return n / 2 }
	if got := backoffDelay(1, noJitter); got != 250*time.Millisecond {
		t.Fatalf("retry 1 = %s", got)
	}
	if got := backoffDelay(2, noJitter); got != 500*time.Millisecond {
		t.Fatalf("retry 2 = %s", got)
	}
	if got := backoffDelay(3, noJitter); got != time.Second {
		t.Fatalf("retry 3 = %s", got)
	}
	if got := backoffDelay(4, noJitter); got != 2*time.Second {
		t.Fatalf("retry 4 = %s", got)
	}
	if got := backoffDelay(20, noJitter); got != retryCap {
		t.Fatalf("cap = %s", got)
	}
}

func TestPostRetriesTransientThenSucceeds(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) < 3 {
			http.Error(w, "busy", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	c := clientFor(t, srv.URL)
	if err := c.SendMetrics(context.Background(), []model.MetricPoint{{MetricName: "cpu.usage"}}); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 3 {
		t.Fatalf("hits = %d", hits.Load())
	}
}

func TestPostDoesNotRetryClientErrors(t *testing.T) {
	for _, code := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden} {
		var hits atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits.Add(1)
			http.Error(w, "no", code)
		}))
		c := clientFor(t, srv.URL)
		err := c.SendMetrics(context.Background(), []model.MetricPoint{{MetricName: "cpu.usage"}})
		srv.Close()
		if err == nil {
			t.Fatalf("status %d: expected error", code)
		}
		var status *hubStatusError
		if !errors.As(err, &status) || status.code != code {
			t.Fatalf("status %d: %v", code, err)
		}
		if hits.Load() != 1 {
			t.Fatalf("status %d hits = %d", code, hits.Load())
		}
	}
}

func TestPostRetriesTooManyRequests(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) == 1 {
			http.Error(w, "slow down", http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()
	c := clientFor(t, srv.URL)
	if err := c.SendEvent(context.Background(), model.Event{Type: "container.die"}); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 2 {
		t.Fatalf("hits = %d", hits.Load())
	}
}

func TestPostStopsWhenContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "down", http.StatusBadGateway)
	}))
	defer srv.Close()
	c := clientFor(t, srv.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := c.SendLogs(ctx, []model.LogEntry{{Message: "x"}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
}

func TestPostGivesUpAfterAttempts(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		http.Error(w, "down", http.StatusBadGateway)
	}))
	defer srv.Close()
	c := clientFor(t, srv.URL)
	err := c.SendFleet(context.Background(), []model.ContainerFleetStatus{{Container: "api"}})
	if err == nil {
		t.Fatal("expected error")
	}
	if hits.Load() != retryAttempts {
		t.Fatalf("hits = %d", hits.Load())
	}
}

func clientFor(t *testing.T, hub string) *Client {
	t.Helper()
	c := NewClient(Config{
		HubURL:     hub,
		HTTPClient: &http.Client{Timeout: time.Second},
	}, nil)
	c.retryWait = func(context.Context, time.Duration) error { return nil }
	return c
}
