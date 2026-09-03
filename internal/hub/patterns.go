package hub

import (
	"net/http"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func (s *Server) registerPatternRoutes() {
	s.mux.HandleFunc("GET /api/v1/logs/patterns", s.handleLogPatterns)
}

func (s *Server) handleLogPatterns(w http.ResponseWriter, r *http.Request) {
	since := time.Now().UTC().Add(-1 * time.Hour)
	if raw := r.URL.Query().Get("since"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			since = time.Now().UTC().Add(-d)
		}
	}
	patterns, err := s.store.ListLogPatterns(r.Context(), since, 50)
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	if patterns == nil {
		patterns = []model.LogPattern{}
	}
	writeJSON(w, http.StatusOK, patterns)
}
