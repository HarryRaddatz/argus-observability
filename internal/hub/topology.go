package hub

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func (s *Server) registerTopologyRoutes() {
	s.mux.HandleFunc("GET /api/v1/topology", s.handleTopology)
	s.mux.HandleFunc("POST /api/v1/topology/batch", s.ingest(s.handleTopologyBatch))
	s.mux.HandleFunc("GET /api/v1/alerts/active", s.handleActiveAlerts)
}

// handleTopologyBatch stores dependencies observed by an agent. These carry the
// kernel origin and take precedence over edges inferred from log text.
func (s *Server) handleTopologyBatch(w http.ResponseWriter, r *http.Request) {
	var links []model.TopologyLink
	if err := json.NewDecoder(r.Body).Decode(&links); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(links) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	now := time.Now().UTC()
	for i := range links {
		if links[i].TS.IsZero() {
			links[i].TS = now
		}
		if links[i].Kind == "" {
			links[i].Kind = "tcp"
		}
	}
	if err := s.store.RecordTopologyLinks(r.Context(), links); err != nil {
		s.logger.Error("record topology links", "err", err)
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) handleTopology(w http.ResponseWriter, r *http.Request) {
	since := time.Now().UTC().Add(-24 * time.Hour)
	if raw := r.URL.Query().Get("since"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			since = time.Now().UTC().Add(-d)
		}
	}
	graph, err := s.store.GetTopology(r.Context(), since)
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	if graph.Nodes == nil {
		graph.Nodes = []model.TopologyNode{}
	}
	if graph.Edges == nil {
		graph.Edges = []model.TopologyEdge{}
	}
	writeJSON(w, http.StatusOK, graph)
}

func (s *Server) handleActiveAlerts(w http.ResponseWriter, _ *http.Request) {
	if s.rules == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	alerts := s.rules.ActiveAlerts()
	out := make([]map[string]any, 0, len(alerts))
	for _, a := range alerts {
		out = append(out, map[string]any{
			"rule_id": a.RuleID, "entity_uid": a.EntityUID, "container": a.Container,
			"title": a.Title, "severity": a.Severity, "summary": a.Summary,
			"fired_at": a.FiredAt.Format(time.RFC3339), "value": a.Value,
		})
	}
	writeJSON(w, http.StatusOK, out)
}
