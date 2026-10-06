package hub

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/bus"
	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/HarryRaddatz/argus-observability/internal/rules"
	"github.com/HarryRaddatz/argus-observability/internal/slo"
	"github.com/HarryRaddatz/argus-observability/internal/store"
	"github.com/HarryRaddatz/argus-observability/internal/store/postgres"
)

func testServer(t *testing.T, token string) (*Server, store.Store) {
	t.Helper()
	dsn := os.Getenv("ARGUS_STORE_DSN")
	if dsn == "" {
		t.Skip("ARGUS_STORE_DSN is not set")
	}
	st, err := postgres.OpenIsolated(dsn, postgres.SchemaName("hub_"+t.Name()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	eventBus := bus.New()
	s := &Server{
		cfg: Config{
			AgentToken:        token,
			IngestConcurrency: 4,
			IngestWait:        time.Second,
			MaxBodyBytes:      1 << 20,
		},
		store:        st,
		bus:          eventBus,
		logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		mux:          http.NewServeMux(),
		ingestSlots:  make(chan struct{}, 4),
		patternQueue: make(chan []model.LogEntry, 8),
		rules:        rules.NewEngine(eventBus),
		sloEval:      slo.NewEvaluator(eventBus),
		staleAgents:  map[string]bool{},
		lastAlert:    map[string]time.Time{},
	}
	s.routes()
	return s, st
}

func doJSON(t *testing.T, h http.Handler, method, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		rdr = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rdr)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	s, _ := testServer(t, "secret")
	rec := doJSON(t, s.Handler(), http.MethodGet, "/health", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["status"] != "ok" {
		t.Fatalf("got %+v", got)
	}
}

func TestMetricsBatchRequiresToken(t *testing.T) {
	s, _ := testServer(t, "secret")
	rec := doJSON(t, s.Handler(), http.MethodPost, "/api/v1/metrics/batch", []model.MetricPoint{}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestMetricsBatchInvalidJSON(t *testing.T) {
	s, _ := testServer(t, "secret")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/metrics/batch", bytes.NewReader([]byte("{")))
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestMetricsBatchAndQueries(t *testing.T) {
	s, _ := testServer(t, "secret")
	now := time.Now().UTC().Truncate(time.Second)
	points := []model.MetricPoint{{
		MetricName: "cpu.usage",
		TS:         now,
		Value:      12.5,
		EntityUID:  "docker:host:api",
		Labels:     model.Labels{"host": "host", "container": "api"},
	}}
	auth := map[string]string{"Authorization": "Bearer secret"}
	rec := doJSON(t, s.Handler(), http.MethodPost, "/api/v1/metrics/batch", points, auth)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("ingest status %d body %s", rec.Code, rec.Body.String())
	}

	work := doJSON(t, s.Handler(), http.MethodGet, "/api/v1/workloads", nil, nil)
	if work.Code != http.StatusOK {
		t.Fatalf("workloads %d %s", work.Code, work.Body.String())
	}

	series := doJSON(t, s.Handler(), http.MethodGet, "/api/v1/metrics/series?metric=cpu.usage&since=2h", nil, nil)
	if series.Code != http.StatusOK {
		t.Fatalf("series %d %s", series.Code, series.Body.String())
	}

	logs := doJSON(t, s.Handler(), http.MethodGet, "/api/v1/logs/search?since=1h", nil, nil)
	if logs.Code != http.StatusOK {
		t.Fatalf("logs %d %s", logs.Code, logs.Body.String())
	}

	fleet := doJSON(t, s.Handler(), http.MethodGet, "/api/v1/fleet/status", nil, nil)
	if fleet.Code != http.StatusOK {
		t.Fatalf("fleet %d %s", fleet.Code, fleet.Body.String())
	}
}

func TestLogsSearchAfterIngest(t *testing.T) {
	s, _ := testServer(t, "secret")
	now := time.Now().UTC()
	entries := []model.LogEntry{{
		TS:        now,
		Message:   "ready",
		Level:     "info",
		EntityUID: "docker:host:api",
		Labels:    model.Labels{"host": "host", "container": "api"},
	}}
	auth := map[string]string{"Authorization": "Bearer secret"}
	rec := doJSON(t, s.Handler(), http.MethodPost, "/api/v1/logs/batch", entries, auth)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("ingest %d %s", rec.Code, rec.Body.String())
	}
	search := doJSON(t, s.Handler(), http.MethodGet, "/api/v1/logs/search?q=ready&since=1h", nil, nil)
	if search.Code != http.StatusOK {
		t.Fatalf("search %d %s", search.Code, search.Body.String())
	}
}
