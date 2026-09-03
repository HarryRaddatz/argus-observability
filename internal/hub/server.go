package hub

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/bus"
	"github.com/HarryRaddatz/argus-observability/internal/insights"
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

	staleMu    sync.Mutex
	staleAgents map[string]bool
	alertMu    sync.Mutex
	lastAlert  map[string]time.Time
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
	s := &Server{
		cfg: cfg, store: st, bus: eventBus, logger: logger, mux: http.NewServeMux(),
		staleAgents: map[string]bool{},
		lastAlert:   map[string]time.Time{},
	}
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
	s.mux.HandleFunc("GET /api/v1/metrics/series", s.handleMetricSeries)
	s.mux.HandleFunc("GET /api/v1/workloads", s.handleWorkloads)
	s.mux.HandleFunc("GET /api/v1/events", s.handleListEvents)
	s.mux.HandleFunc("GET /api/v1/logs/search", s.handleSearchLogs)
	s.mux.HandleFunc("GET /api/v1/insights", s.handleInsights)
	s.mux.HandleFunc("GET /api/v1/metrics/catalog", s.handleMetricsCatalog)
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
	s.staleMu.Lock()
	delete(s.staleAgents, req.AgentID)
	s.staleMu.Unlock()
	s.bus.Publish(model.Event{
		ID: uuid.NewString(), Type: "agent.register", TS: now,
		Severity: "info", Source: "hub",
		EntityUID: "host:" + req.HostID,
		Labels:    model.Labels{"host": req.HostID, "agent_id": req.AgentID, "runtime": req.Runtime},
	})
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
	s.staleMu.Lock()
	wasStale := s.staleAgents[req.AgentID]
	if wasStale {
		delete(s.staleAgents, req.AgentID)
	}
	s.staleMu.Unlock()
	if wasStale {
		hostID := req.AgentID
		if reg, err := s.store.GetAgent(r.Context(), req.AgentID); err == nil && reg.HostID != "" {
			hostID = reg.HostID
		}
		s.bus.Publish(model.Event{
			ID: uuid.NewString(), Type: "agent.reconnect", TS: time.Now().UTC(),
			Severity: "info", Source: "hub",
			EntityUID: "host:" + hostID,
			Labels:    model.Labels{"host": hostID, "agent_id": req.AgentID},
		})
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
	points = append(points, deriveMemoryPct(points)...)
	if err := s.store.WriteMetrics(r.Context(), points); err != nil {
		s.logger.Error("write metrics", "err", err)
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	s.checkResourcePressure(points)
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
	for i := range entries {
		_, fields := insights.EnrichLog(entries[i].Message, entries[i].Level)
		entries[i].Fields = fields
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

func (s *Server) handleMetricSeries(w http.ResponseWriter, r *http.Request) {
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		http.Error(w, "metric required", http.StatusBadRequest)
		return
	}
	container := r.URL.Query().Get("container")
	since := time.Now().UTC().Add(-1 * time.Hour)
	if raw := r.URL.Query().Get("since"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			since = time.Now().UTC().Add(-d)
		}
	}
	series, err := s.store.QueryMetricSeries(r.Context(), metric, container, since)
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	if series == nil {
		series = []model.ContainerSeries{}
	}
	writeJSON(w, http.StatusOK, model.MetricSeriesResponse{MetricName: metric, Series: series})
}

func (s *Server) handleWorkloads(w http.ResponseWriter, r *http.Request) {
	since := time.Now().UTC().Add(-15 * time.Minute)
	if raw := r.URL.Query().Get("since"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			since = time.Now().UTC().Add(-d)
		}
	}
	workloads, err := s.store.ListWorkloads(r.Context(), since)
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	if workloads == nil {
		workloads = []model.WorkloadSnapshot{}
	}
	writeJSON(w, http.StatusOK, workloads)
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
	if events == nil {
		events = []model.Event{}
	}
	writeJSON(w, http.StatusOK, events)
}

func (s *Server) handleSearchLogs(w http.ResponseWriter, r *http.Request) {
	since := time.Now().UTC().Add(-15 * time.Minute)
	if raw := r.URL.Query().Get("since"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			since = time.Now().UTC().Add(-d)
		}
	}
	filter := model.LogSearchFilter{
		Query:     r.URL.Query().Get("q"),
		EntityUID: r.URL.Query().Get("entity_uid"),
		Container: r.URL.Query().Get("container"),
		Level:     r.URL.Query().Get("level"),
		Topic:     r.URL.Query().Get("topic"),
		Since:     since,
		Limit:     200,
	}
	logs, err := s.store.SearchLogs(r.Context(), filter)
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	if logs == nil {
		logs = []model.LogEntry{}
	}
	writeJSON(w, http.StatusOK, logs)
}

func (s *Server) handleInsights(w http.ResponseWriter, r *http.Request) {
	since := time.Now().UTC().Add(-1 * time.Hour)
	if raw := r.URL.Query().Get("since"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			since = time.Now().UTC().Add(-d)
		}
	}
	workloads, err := s.store.ListWorkloads(r.Context(), since)
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	var wlMem []insights.WorkloadMemory
	for _, w := range workloads {
		pct := 0.0
		if w.MemoryLimit > 0 {
			pct = (w.MemoryUsage / w.MemoryLimit) * 100
		}
		wlMem = append(wlMem, insights.WorkloadMemory{
			Container: w.Container, EntityUID: w.EntityUID,
			CPUUsage: w.CPUUsage, MemoryPct: pct,
			MemoryUsage: w.MemoryUsage, MemoryLimit: w.MemoryLimit,
		})
	}
	logStatsRaw, err := s.store.CountLogTopics(r.Context(), since)
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	var logStats []insights.LogTopicStats
	for _, ls := range logStatsRaw {
		logStats = append(logStats, insights.LogTopicStats{
			Container: ls.Container, EntityUID: ls.EntityUID, Topic: ls.Topic, Count: ls.Count,
		})
	}
	result := model.InsightsResponse{
		Since:    since.Format(time.RFC3339),
		Insights: insights.Generate(wlMem, logStats),
	}
	if result.Insights == nil {
		result.Insights = []model.Insight{}
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleMetricsCatalog(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, []map[string]string{
		{"name": "cpu.usage", "label": "CPU %", "unit": "%"},
		{"name": "memory.usage", "label": "Memória (bytes)", "unit": "bytes"},
		{"name": "memory.usage_pct", "label": "Memória %", "unit": "%"},
		{"name": "memory.limit", "label": "Limite memória", "unit": "bytes"},
	})
}

func deriveMemoryPct(points []model.MetricPoint) []model.MetricPoint {
	type snap struct {
		use float64
		lim float64
		ts  time.Time
		uid string
		lbl model.Labels
	}
	byEntity := map[string]*snap{}
	for _, p := range points {
		em, ok := byEntity[p.EntityUID]
		if !ok {
			em = &snap{uid: p.EntityUID, lbl: p.Labels, ts: p.TS}
			byEntity[p.EntityUID] = em
		}
		switch p.MetricName {
		case "memory.usage":
			em.use = p.Value
			em.ts = p.TS
		case "memory.limit":
			em.lim = p.Value
		}
	}
	var out []model.MetricPoint
	for _, em := range byEntity {
		if em.lim <= 0 {
			continue
		}
		out = append(out, model.MetricPoint{
			MetricName: "memory.usage_pct",
			TS:         em.ts,
			Value:      (em.use / em.lim) * 100,
			EntityUID:  em.uid,
			Labels:     em.lbl,
		})
	}
	return out
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
			s.staleMu.Lock()
			if s.staleAgents[a.AgentID] {
				s.staleMu.Unlock()
				continue
			}
			s.staleAgents[a.AgentID] = true
			s.staleMu.Unlock()
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

func (s *Server) checkResourcePressure(points []model.MetricPoint) {
	type entityMetrics struct {
		cpu    float64
		memUse float64
		memLim float64
		labels model.Labels
	}
	byEntity := map[string]*entityMetrics{}
	for _, p := range points {
		em, ok := byEntity[p.EntityUID]
		if !ok {
			em = &entityMetrics{labels: p.Labels}
			byEntity[p.EntityUID] = em
		}
		switch p.MetricName {
		case "cpu.usage":
			em.cpu = p.Value
		case "memory.usage":
			em.memUse = p.Value
		case "memory.limit":
			em.memLim = p.Value
		}
	}
	now := time.Now().UTC()
	for entityUID, em := range byEntity {
		if em.cpu > 80 {
			s.maybeAlert(entityUID+":cpu", now, model.Event{
				ID: uuid.NewString(), Type: "resource.pressure", TS: now,
				Severity: "warning", Source: "hub", EntityUID: entityUID, Labels: em.labels,
				Payload: map[string]any{"metric": "cpu.usage", "value": em.cpu, "threshold": 80},
			})
		}
		if em.memLim > 0 {
			pct := (em.memUse / em.memLim) * 100
			if pct > 90 {
				s.maybeAlert(entityUID+":mem", now, model.Event{
					ID: uuid.NewString(), Type: "resource.pressure", TS: now,
					Severity: "warning", Source: "hub", EntityUID: entityUID, Labels: em.labels,
					Payload: map[string]any{"metric": "memory.usage", "value_pct": pct, "threshold": 90},
				})
			}
		}
	}
}

func (s *Server) maybeAlert(key string, now time.Time, evt model.Event) {
	s.alertMu.Lock()
	defer s.alertMu.Unlock()
	if last, ok := s.lastAlert[key]; ok && now.Sub(last) < 5*time.Minute {
		return
	}
	s.lastAlert[key] = now
	s.bus.Publish(evt)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

var ErrUnauthorized = errors.New("unauthorized")
