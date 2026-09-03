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
	GetAgent(ctx context.Context, agentID string) (model.AgentRegistration, error)
	StaleAgents(ctx context.Context, before time.Time) ([]model.AgentRegistration, error)

	WriteMetrics(ctx context.Context, points []model.MetricPoint) error
	WriteLogs(ctx context.Context, entries []model.LogEntry) error
	WriteEvents(ctx context.Context, events []model.Event) error

	QueryMetrics(ctx context.Context, metricName string, labels model.Labels, since time.Time) ([]model.SeriesPoint, error)
	QueryMetricSeries(ctx context.Context, metricName, container string, since time.Time) ([]model.ContainerSeries, error)
	QueryHTTPServiceSummary(ctx context.Context, since time.Time) ([]model.HTTPServiceSummary, error)
	ListWorkloads(ctx context.Context, since time.Time) ([]model.WorkloadSnapshot, error)
	ListEvents(ctx context.Context, entityUID string, since time.Time, limit int) ([]model.Event, error)
	SearchLogs(ctx context.Context, filter model.LogSearchFilter) ([]model.LogEntry, error)
	CountLogTopics(ctx context.Context, since time.Time) ([]model.LogTopicCount, error)

	UpsertFleetStatus(ctx context.Context, rows []model.ContainerFleetStatus) error
	GetFleetStatus(ctx context.Context) ([]model.ContainerFleetStatus, error)
	CountFleetEvents(ctx context.Context, since time.Time) (model.FleetEventStats, error)

	ListWorkloadGroups(ctx context.Context) ([]model.WorkloadGroup, error)
	GetWorkloadGroup(ctx context.Context, id string) (model.WorkloadGroup, error)
	CreateWorkloadGroup(ctx context.Context, in model.WorkloadGroupInput) (model.WorkloadGroup, error)
	UpdateWorkloadGroup(ctx context.Context, id string, in model.WorkloadGroupInput) (model.WorkloadGroup, error)
	DeleteWorkloadGroup(ctx context.Context, id string) error

	Purge(ctx context.Context, logsBefore, metricsBefore, eventsBefore time.Time) (model.PurgeResult, error)

	RecordLogPatterns(ctx context.Context, entries []model.LogEntry) error
	ListLogPatterns(ctx context.Context, since time.Time, limit int) ([]model.LogPattern, error)
	RecordTopologyEdges(ctx context.Context, entries []model.LogEntry) error
	GetTopology(ctx context.Context, since time.Time) (model.TopologyGraph, error)
}
