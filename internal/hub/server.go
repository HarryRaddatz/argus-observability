package hub

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/bus"
	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/HarryRaddatz/argus-observability/internal/store"
	"github.com/google/uuid"
)

type Config struct {
	Addr        string
	AgentToken  string
	CORSOrigin  string
	StaleAfter  time.Duration
	Heartbeat   time.Duration
}

type Server struct {
	cfg    Config
	store  store.Store
	bus    *bus.Bus
	logger *slog.Logger
	mux    *http.ServeMux
}

func New(cfg Config, st store.Store, eventBus *bus.Bus, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	if cfg.StaleAfter == 0 {
		cfg.StaleAfter = 2 * time.Minute
	}
	if cfg.Heartbeat == 0 {
		cfg.Heartbeat = 30 * time.Second
	}
	s := &Server{cfg: cfg, store: st, bus: eventBus, logger: logger, mux: http.NewServeMux()}
	s.routes()
	eventBus.Subscribe(s.onEvent)
	go s.staleLoop()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("POST /api/v1/agents/register", s.auth(s.handleRegister))
	s.mux.HandleFunc("POST /api/v1/agents/heartbeat", s.auth(s.handleHeartbeat))
	s.mux.HandleFunc("POST /api/v1/metrics/batch", s.auth(s.handleMetricsBatch))
	s.mux.HandleFunc("POST /api/v1/logs/batch", s.auth(s.handleLogsBatch))
	s.mux.HandleFunc("POST /api/v1/events", s.auth(s.handleEventIngest))
	s.mux.HandleFunc("GET /api/v1/query", s.handleQuery)
	s.mux.HandleFunc("GET /api/v1/events", s.handleListEvents)
	s.mux.HandleFunc("GET /api/v1/logs/search", s.handleSearchLogs)
}

func (s *Server) Handler() http.Handler {
	return withCORS(s.cfg.CORSOrigin, s.mux)
}

func withCORS(origin string, next http.Handler) http.Handler {
	if origin == "" {
		origin = "*"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.AgentToken == "" {
			next(w, r)
			return
		}
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token != s.cfg.AgentToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type registerRequest struct {
	AgentID string       `json:"agent_id"`
	HostID  string       `json:"host_id"`
	Runtime string       `json:"runtime"`
	Labels  model.Labels `json:"labels"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.AgentID == "" || req.HostID == "" {
		http.Error(w, "agent_id and host_id required", http.StatusBadRequest)
		return
	}
	if req.Runtime == "" {
		req.Runtime = "docker"
	}
	now := time.Now().UTC()
	if err := s.store.UpsertAgent(r.Context(), model.AgentRegistration{
		AgentID: req.AgentID, HostID: req.HostID, Runtime: req.Runtime, Labels: req.Labels, LastSeen: now,
	}); err != nil {
		s.logger.Error("register agent", "err", err)
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, model.AgentSession{
		SessionID: uuid.NewString(),
		Interval:  int(s.cfg.Heartbeat.Seconds()),
	})
}

type heartbeatRequest struct {
	AgentID string `json:"agent_id"`
}

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	var req heartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.AgentID == "" {
		http.Error(w, "agent_id required", http.StatusBadRequest)
		return
	}
	if err := s.store.TouchAgent(r.Context(), req.AgentID, time.Now().UTC()); err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleMetricsBatch(w http.ResponseWriter, r *http.Request) {
	var points []model.MetricPoint
	if err := json.NewDecoder(r.Body).Decode(&points); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(points) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	if err := s.store.WriteMetrics(r.Context(), points); err != nil {
		s.logger.Error("write metrics", "err", err)
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) handleLogsBatch(w http.ResponseWriter, r *http.Request) {
	var entries []model.LogEntry
	if err := json.NewDecoder(r.Body).Decode(&entries); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(entries) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	if err := s.store.WriteLogs(r.Context(), entries); err != nil {
		s.logger.Error("write logs", "err", err)
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) handleEventIngest(w http.ResponseWriter, r *http.Request) {
	var evt model.Event
	if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if evt.ID == "" {
		evt.ID = uuid.NewString()
	}
	if evt.TS.IsZero() {
		evt.TS = time.Now().UTC()
	}
	s.bus.Publish(evt)
	writeJSON(w, http.StatusAccepted, map[string]string{"id": evt.ID})
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		http.Error(w, "metric required", http.StatusBadRequest)
		return
	}
	since := time.Now().UTC().Add(-1 * time.Hour)
	if raw := r.URL.Query().Get("since"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			since = time.Now().UTC().Add(-d)
		}
	}
	points, err := s.store.QueryMetrics(r.Context(), metric, nil, since)
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, model.QuerySeries{MetricName: metric, Points: points})
}

func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	entityUID := r.URL.Query().Get("entity_uid")
	since := time.Now().UTC().Add(-24 * time.Hour)
	if raw := r.URL.Query().Get("since"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			since = time.Now().UTC().Add(-d)
		}
	}
	events, err := s.store.ListEvents(r.Context(), entityUID, since, 200)
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (s *Server) handleSearchLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	entityUID := r.URL.Query().Get("entity_uid")
	since := time.Now().UTC().Add(-15 * time.Minute)
	if raw := r.URL.Query().Get("since"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			since = time.Now().UTC().Add(-d)
		}
	}
	logs, err := s.store.SearchLogs(r.Context(), q, entityUID, since, 200)
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

func (s *Server) onEvent(evt model.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.store.WriteEvents(ctx, []model.Event{evt}); err != nil {
		s.logger.Error("persist event", "type", evt.Type, "err", err)
	}
}

func (s *Server) staleLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	seen := map[string]struct{}{}
	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cutoff := time.Now().UTC().Add(-s.cfg.StaleAfter)
		stale, err := s.store.StaleAgents(ctx, cutoff)
		cancel()
		if err != nil {
			s.logger.Error("stale agents", "err", err)
			continue
		}
		for _, a := range stale {
			if _, ok := seen[a.AgentID]; ok {
				continue
			}
			seen[a.AgentID] = struct{}{}
			s.bus.Publish(model.Event{
				ID: uuid.NewString(), Type: "agent.disconnect", TS: time.Now().UTC(),
				Severity: "warning", Source: "hub",
				EntityUID: "host:" + a.HostID,
				Labels:    model.Labels{"host": a.HostID, "agent_id": a.AgentID},
				Payload:   map[string]any{"last_seen": a.LastSeen},
			})
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

var ErrUnauthorized = errors.New("unauthorized")
