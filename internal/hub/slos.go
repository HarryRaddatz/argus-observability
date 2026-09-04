package hub

import (
	"net/http"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func (s *Server) registerSLORoutes() {
	s.mux.HandleFunc("GET /api/v1/slos", s.handleListSLOs)
	s.mux.HandleFunc("GET /api/v1/slos/status", s.handleListSLOStatuses)
	s.mux.HandleFunc("GET /api/v1/slos/{id}/status", s.handleSLOStatus)
}

func (s *Server) handleListSLOStatuses(w http.ResponseWriter, r *http.Request) {
	defs, err := s.store.ListSLOs(r.Context())
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	now := time.Now().UTC()
	out := make([]model.SLOStatus, 0, len(defs))
	for _, def := range defs {
		status, err := s.store.EvaluateSLO(r.Context(), def, now)
		if err != nil {
			continue
		}
		out = append(out, status)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleListSLOs(w http.ResponseWriter, r *http.Request) {
	slos, err := s.store.ListSLOs(r.Context())
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	if slos == nil {
		slos = []model.SLODefinition{}
	}
	writeJSON(w, http.StatusOK, slos)
}

func (s *Server) handleSLOStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	def, err := s.store.GetSLO(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	status, err := s.store.EvaluateSLO(r.Context(), def, time.Now().UTC())
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, status)
}
