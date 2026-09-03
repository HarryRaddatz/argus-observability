package hub

import (
	"context"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/google/uuid"
)

func (s *Server) retentionLoop() {
	if s.cfg.RetentionLogs <= 0 && s.cfg.RetentionMetrics <= 0 && s.cfg.RetentionEvents <= 0 {
		return
	}
	interval := s.cfg.PurgeInterval
	if interval <= 0 {
		interval = time.Hour
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	s.runPurge()
	for range ticker.C {
		go s.runPurge()
	}
}

func (s *Server) runPurge() {
	timeout := s.cfg.PurgeTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	now := time.Now().UTC()
	logsBefore := now.Add(-s.cfg.RetentionLogs)
	metricsBefore := now.Add(-s.cfg.RetentionMetrics)
	eventsBefore := now.Add(-s.cfg.RetentionEvents)

	result, err := s.store.Purge(ctx, logsBefore, metricsBefore, eventsBefore)
	if err != nil && ctx.Err() == nil {
		s.logger.Error("retention purge", "err", err)
		return
	}

	s.logger.Info("retention purge",
		"logs_deleted", result.LogsDeleted,
		"metrics_deleted", result.MetricsDeleted,
		"events_deleted", result.EventsDeleted,
		"duration_ms", result.Duration.Milliseconds(),
		"truncated", result.Truncated,
	)

	if result.LogsDeleted == 0 && result.MetricsDeleted == 0 && result.EventsDeleted == 0 {
		return
	}

	s.bus.Publish(model.Event{
		ID:        uuid.NewString(),
		Type:      "retention.purge",
		TS:        now,
		Severity:  "info",
		Source:    "hub",
		EntityUID: "hub:retention",
		Labels:    model.Labels{"component": "retention"},
		Payload: map[string]any{
			"logs_deleted":    result.LogsDeleted,
			"metrics_deleted": result.MetricsDeleted,
			"events_deleted":  result.EventsDeleted,
			"duration_ms":     result.Duration.Milliseconds(),
			"truncated":       result.Truncated,
		},
	})
}
