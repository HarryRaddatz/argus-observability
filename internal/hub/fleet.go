package hub

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/fleet"
	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func (s *Server) registerFleetRoutes() {
	s.mux.HandleFunc("POST /api/v1/fleet/batch", s.auth(s.handleFleetBatch))
	s.mux.HandleFunc("GET /api/v1/fleet/status", s.handleFleetStatus)
}

func (s *Server) handleFleetBatch(w http.ResponseWriter, r *http.Request) {
	var rows []model.ContainerFleetStatus
	if err := json.NewDecoder(r.Body).Decode(&rows); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(rows) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	if err := s.store.UpsertFleetStatus(r.Context(), rows); err != nil {
		s.logger.Error("upsert fleet", "err", err)
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) handleFleetStatus(w http.ResponseWriter, r *http.Request) {
	containers, err := s.store.GetFleetStatus(r.Context())
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	events, err := s.store.CountFleetEvents(r.Context(), time.Now().UTC().Add(-24*time.Hour))
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	resp := fleet.BuildResponse(containers, events)
	writeJSON(w, http.StatusOK, resp)
}
