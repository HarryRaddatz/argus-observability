package hub

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/groups"
	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/HarryRaddatz/argus-observability/internal/store/sqlite"
)

func (s *Server) registerGroupRoutes() {
	s.mux.HandleFunc("GET /api/v1/workload-groups", s.handleListGroups)
	s.mux.HandleFunc("GET /api/v1/workload-groups/discover", s.handleDiscoverGroups)
	s.mux.HandleFunc("POST /api/v1/workload-groups", s.handleCreateGroup)
	s.mux.HandleFunc("GET /api/v1/workload-groups/{id}", s.handleGetGroup)
	s.mux.HandleFunc("GET /api/v1/workload-groups/{id}/summary", s.handleGroupSummary)
	s.mux.HandleFunc("PUT /api/v1/workload-groups/{id}", s.handleUpdateGroup)
	s.mux.HandleFunc("DELETE /api/v1/workload-groups/{id}", s.handleDeleteGroup)
}

func (s *Server) handleListGroups(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListWorkloadGroups(r.Context())
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []model.WorkloadGroup{}
	}
	workloads, _ := s.store.ListWorkloads(r.Context(), time.Now().UTC().Add(-30*time.Minute))
	for i := range list {
		list[i].MemberCount = len(groups.FilterWorkloads(workloads, list[i]))
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleDiscoverGroups(w http.ResponseWriter, r *http.Request) {
	since := time.Now().UTC().Add(-30 * time.Minute)
	workloads, err := s.store.ListWorkloads(r.Context(), since)
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	discovered := groups.Discover(workloads)
	if discovered == nil {
		discovered = []model.WorkloadGroup{}
	}
	writeJSON(w, http.StatusOK, discovered)
}

func (s *Server) handleCreateGroup(w http.ResponseWriter, r *http.Request) {
	var in model.WorkloadGroupInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	in = normalizeGroupInput(in)
	if err := validateGroupInput(in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	g, err := s.store.CreateWorkloadGroup(r.Context(), in)
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, g)
}

func (s *Server) handleGetGroup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	g, err := s.store.GetWorkloadGroup(r.Context(), id)
	if err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, g)
}

func (s *Server) handleUpdateGroup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in model.WorkloadGroupInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	in = normalizeGroupInput(in)
	if err := validateGroupInput(in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	g, err := s.store.UpdateWorkloadGroup(r.Context(), id, in)
	if err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, g)
}

func (s *Server) handleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.DeleteWorkloadGroup(r.Context(), id); err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGroupSummary(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	g, err := s.store.GetWorkloadGroup(r.Context(), id)
	if err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	since := time.Now().UTC().Add(-30 * time.Minute)
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
	members := groups.FilterWorkloads(workloads, g)
	summary := groups.Summarize(g, members)
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) resolveGroupContainers(ctx context.Context, groupID string) ([]string, error) {
	if groupID == "" {
		return nil, nil
	}
	g, err := s.store.GetWorkloadGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	workloads, err := s.store.ListWorkloads(ctx, time.Now().UTC().Add(-1*time.Hour))
	if err != nil {
		return nil, err
	}
	members := groups.FilterWorkloads(workloads, g)
	names := make([]string, 0, len(members))
	for _, m := range members {
		names = append(names, m.Container)
	}
	return names, nil
}

func normalizeGroupInput(in model.WorkloadGroupInput) model.WorkloadGroupInput {
	in.Name = strings.TrimSpace(in.Name)
	in.LabelValue = strings.TrimSpace(in.LabelValue)
	if in.LabelKey == "" && (in.Kind == model.GroupKindStack || in.Kind == model.GroupKindService) {
		in.LabelKey = in.Kind
	}
	return in
}

func validateGroupInput(in model.WorkloadGroupInput) error {
	if in.Name == "" {
		return errors.New("name required")
	}
	switch in.Kind {
	case model.GroupKindStack, model.GroupKindService:
		if in.LabelValue == "" {
			return errors.New("label_value required for stack/service groups")
		}
	case model.GroupKindCustom:
		if len(in.Containers) == 0 {
			return errors.New("containers required for custom group")
		}
	default:
		return errors.New("kind must be stack, service, or custom")
	}
	return nil
}
