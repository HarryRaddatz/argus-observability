package hub

import (
	"net/http"
	"strings"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/HarryRaddatz/argus-observability/internal/traces"
)

func (s *Server) registerTraceRoutes() {
	s.mux.HandleFunc("GET /api/v1/traces/{trace_id}", s.handleGetTrace)
}

func (s *Server) handleGetTrace(w http.ResponseWriter, r *http.Request) {
	traceID := strings.TrimSpace(r.PathValue("trace_id"))
	if traceID == "" {
		http.Error(w, "trace_id required", http.StatusBadRequest)
		return
	}

	stored, err := s.store.GetTraceSpans(r.Context(), traceID)
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	if len(stored) > 0 {
		detail := buildTraceDetail(traceID, "otlp", stored)
		writeJSON(w, http.StatusOK, detail)
		return
	}

	since := time.Now().UTC().Add(-24 * time.Hour)
	if raw := r.URL.Query().Get("since"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			since = time.Now().UTC().Add(-d)
		}
	}
	logs, err := s.store.SearchLogs(r.Context(), model.LogSearchFilter{
		TraceID: traceID,
		Since:   since,
		Limit:   500,
	})
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	detail := traces.BuildFromLogs(traceID, logs)
	if detail.Spans == nil {
		detail.Spans = []model.TraceSpan{}
	}
	writeJSON(w, http.StatusOK, detail)
}

func buildTraceDetail(traceID, source string, spans []model.TraceSpan) model.TraceDetail {
	detail := model.TraceDetail{TraceID: traceID, Source: source, Spans: spans}
	if len(spans) == 0 {
		return detail
	}
	detail.StartTS = spans[0].StartTS
	detail.EndTS = spans[len(spans)-1].EndTS
	for _, sp := range spans {
		if sp.StartTS.Before(detail.StartTS) || detail.StartTS.IsZero() {
			detail.StartTS = sp.StartTS
		}
		if sp.EndTS.After(detail.EndTS) {
			detail.EndTS = sp.EndTS
		}
	}
	if detail.EndTS.After(detail.StartTS) {
		detail.DurationMs = float64(detail.EndTS.Sub(detail.StartTS).Milliseconds())
	}
	return detail
}
