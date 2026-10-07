package hub

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/HarryRaddatz/argus-observability/internal/store"
)

const patternLimit = 50

func (s *Server) registerPatternRoutes() {
	s.mux.HandleFunc("GET /api/v1/logs/patterns", s.auth(s.handleLogPatterns))
}

func (s *Server) handleLogPatterns(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	since, _, winErr := requestWindow(r, time.Hour)
	if winErr != nil {
		http.Error(w, "invalid range", http.StatusBadRequest)
		return
	}
	container := strings.TrimSpace(q.Get("container"))
	if container == "all" {
		container = ""
	}
	text := strings.ToLower(strings.TrimSpace(q.Get("q")))
	groupID := strings.TrimSpace(q.Get("group"))

	var allowed map[string]struct{}
	if groupID != "" {
		names, err := s.resolveGroupContainers(r.Context(), groupID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				http.Error(w, "group not found", http.StatusNotFound)
				return
			}
			http.Error(w, "store error", http.StatusInternalServerError)
			return
		}
		allowed = make(map[string]struct{}, len(names))
		for _, n := range names {
			allowed[n] = struct{}{}
		}
	}

	fetch := patternLimit
	filtered := container != "" || text != "" || allowed != nil
	if filtered {
		// Filters apply after the query, so read a wider slice to still fill the page.
		fetch = patternLimit * 20
	}
	patterns, err := s.store.ListLogPatterns(r.Context(), since, fetch)
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}

	out := make([]model.LogPattern, 0, patternLimit)
	for _, p := range patterns {
		if container != "" && p.Container != container {
			continue
		}
		if allowed != nil {
			if _, ok := allowed[p.Container]; !ok {
				continue
			}
		}
		if text != "" && !strings.Contains(strings.ToLower(p.Pattern), text) && !strings.Contains(strings.ToLower(p.Sample), text) {
			continue
		}
		out = append(out, p)
		if len(out) == patternLimit {
			break
		}
	}
	writeJSON(w, http.StatusOK, out)
}
