package store

import (
	"context"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

type Store interface {
	Close() error

	UpsertAgent(ctx context.Context, reg model.AgentRegistration) error
	TouchAgent(ctx context.Context, agentID string, at time.Time) error
	StaleAgents(ctx context.Context, before time.Time) ([]model.AgentRegistration, error)

	WriteMetrics(ctx context.Context, points []model.MetricPoint) error
	WriteLogs(ctx context.Context, entries []model.LogEntry) error
	WriteEvents(ctx context.Context, events []model.Event) error

	QueryMetrics(ctx context.Context, metricName string, labels model.Labels, since time.Time) ([]model.SeriesPoint, error)
	ListEvents(ctx context.Context, entityUID string, since time.Time, limit int) ([]model.Event, error)
	SearchLogs(ctx context.Context, query string, entityUID string, since time.Time, limit int) ([]model.LogEntry, error)
}
