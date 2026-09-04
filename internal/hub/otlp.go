package hub

import (
	"io"
	"net/http"

	"github.com/HarryRaddatz/argus-observability/internal/otel"
)

func (s *Server) registerOTLPRoutes() {
	s.mux.HandleFunc("POST /v1/traces", s.auth(s.handleOTLPTraces))
}

func (s *Server) handleOTLPTraces(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}
	spans, err := otel.ParseOTLPTracesJSON(body)
	if err != nil {
		http.Error(w, "invalid otlp json", http.StatusBadRequest)
		return
	}
	if len(spans) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}
	if err := s.store.WriteTraceSpans(r.Context(), spans); err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
