package sqlite

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/insights"
	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/HarryRaddatz/argus-observability/internal/slo"
)

func (s *SQLite) WriteTraceSpans(ctx context.Context, spans []model.TraceSpan) error {
	if len(spans) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO trace_spans (trace_id, span_id, parent_span_id, name, service, container, entity_uid,
  start_ts, end_ts, duration_ms, status, kind, source, attributes_json)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(trace_id, span_id) DO UPDATE SET
  parent_span_id=excluded.parent_span_id,
  name=excluded.name,
  service=excluded.service,
  container=excluded.container,
  entity_uid=excluded.entity_uid,
  start_ts=excluded.start_ts,
  end_ts=excluded.end_ts,
  duration_ms=excluded.duration_ms,
  status=excluded.status,
  kind=excluded.kind,
  source=excluded.source,
  attributes_json=excluded.attributes_json`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, sp := range spans {
		attrs, err := json.Marshal(sp.Attributes)
		if err != nil {
			return err
		}
		if _, err := stmt.ExecContext(ctx,
			sp.TraceID, sp.SpanID, sp.ParentSpanID, sp.Name, sp.Service, sp.Container, sp.EntityUID,
			sp.StartTS.UTC().Format(time.RFC3339Nano), sp.EndTS.UTC().Format(time.RFC3339Nano),
			sp.DurationMs, sp.Status, sp.Kind, sp.Source, string(attrs),
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLite) GetTraceSpans(ctx context.Context, traceID string) ([]model.TraceSpan, error) {
	norm := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(traceID), "-", ""))
	if norm == "" {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT trace_id, span_id, parent_span_id, name, service, container, entity_uid,
  start_ts, end_ts, duration_ms, status, kind, source, attributes_json
FROM trace_spans
WHERE replace(lower(trace_id), '-', '') = ? OR trace_id LIKE ?
ORDER BY start_ts ASC`, norm, "%"+norm+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTraceSpans(rows)
}

func scanTraceSpans(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]model.TraceSpan, error) {
	out := make([]model.TraceSpan, 0)
	for rows.Next() {
		var sp model.TraceSpan
		var startStr, endStr, attrsJSON string
		if err := rows.Scan(&sp.TraceID, &sp.SpanID, &sp.ParentSpanID, &sp.Name, &sp.Service, &sp.Container,
			&sp.EntityUID, &startStr, &endStr, &sp.DurationMs, &sp.Status, &sp.Kind, &sp.Source, &attrsJSON); err != nil {
			return nil, err
		}
		sp.StartTS, _ = time.Parse(time.RFC3339Nano, startStr)
		sp.EndTS, _ = time.Parse(time.RFC3339Nano, endStr)
		_ = json.Unmarshal([]byte(attrsJSON), &sp.Attributes)
		out = append(out, sp)
	}
	return out, rows.Err()
}

func (s *SQLite) ListSLOs(ctx context.Context) ([]model.SLODefinition, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, name, service, group_id, sli_metric, target, window_hours, latency_threshold_ms, created_at
FROM slos ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.SLODefinition, 0)
	for rows.Next() {
		var def model.SLODefinition
		var groupID, createdAt string
		if err := rows.Scan(&def.ID, &def.Name, &def.Service, &groupID, &def.SLIMetric, &def.Target,
			&def.WindowHours, &def.LatencyThresholdMs, &createdAt); err != nil {
			return nil, err
		}
		def.GroupID = groupID
		def.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		out = append(out, def)
	}
	return out, rows.Err()
}

func (s *SQLite) GetSLO(ctx context.Context, id string) (model.SLODefinition, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, name, service, group_id, sli_metric, target, window_hours, latency_threshold_ms, created_at
FROM slos WHERE id=?`, id)
	var def model.SLODefinition
	var groupID, createdAt string
	if err := row.Scan(&def.ID, &def.Name, &def.Service, &groupID, &def.SLIMetric, &def.Target,
		&def.WindowHours, &def.LatencyThresholdMs, &createdAt); err != nil {
		return model.SLODefinition{}, err
	}
	def.GroupID = groupID
	def.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return def, nil
}

func (s *SQLite) EvaluateSLO(ctx context.Context, def model.SLODefinition, at time.Time) (model.SLOStatus, error) {
	window := time.Duration(def.WindowHours) * time.Hour
	if window <= 0 {
		window = 30 * 24 * time.Hour
	}
	since := at.Add(-window)
	latencies, requests, errors, err := s.httpMetricsForService(ctx, def.Service, since)
	if err != nil {
		return model.SLOStatus{}, err
	}

	status := model.SLOStatus{
		SLO:         def,
		TotalEvents: requests,
		EvaluatedAt: at,
	}
	if def.SLIMetric == "availability" {
		status.GoodEvents = requests - errors
		if requests > 0 {
			status.Compliance = (float64(status.GoodEvents) / float64(requests)) * 100
		} else {
			status.Compliance = 100
		}
	} else {
		status.P95LatencyMs = slo.P95(latencies)
		status.Compliance = slo.ComplianceLatency(latencies, def.LatencyThresholdMs, def.Target)
		status.GoodEvents = 0
		for _, d := range latencies {
			if d <= def.LatencyThresholdMs {
				status.GoodEvents++
			}
		}
		status.TotalEvents = len(latencies)
	}
	status.ErrorBudgetRemaining = slo.ErrorBudgetRemaining(status.Compliance, def.Target)
	status.Breached = status.Compliance < def.Target
	return status, nil
}

func (s *SQLite) httpMetricsForService(ctx context.Context, service string, since time.Time) ([]float64, int, int, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT metric_name, value, labels_json FROM metric_points
WHERE metric_name IN ('http.duration_ms', 'http.requests', 'http.errors') AND ts >= ?
`, since.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	var latencies []float64
	requests, errors := 0, 0
	for rows.Next() {
		var metricName, labelsJSON string
		var value float64
		if err := rows.Scan(&metricName, &value, &labelsJSON); err != nil {
			return nil, 0, 0, err
		}
		var labels model.Labels
		_ = json.Unmarshal([]byte(labelsJSON), &labels)
		svc := labels["service"]
		if svc == "" {
			svc = insights.InferServiceFromContainer(labels["container"])
		}
		if service != "" && svc != service {
			continue
		}
		switch metricName {
		case "http.requests":
			requests += int(value)
		case "http.errors":
			errors += int(value)
		case "http.duration_ms":
			latencies = append(latencies, value)
		}
	}
	return latencies, requests, errors, rows.Err()
}

func (s *SQLite) seedDefaultSLOs() error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.Exec(`
INSERT OR IGNORE INTO slos (id, name, service, group_id, sli_metric, target, window_hours, latency_threshold_ms, created_at)
VALUES
  ('slo-demo-latency', 'Latência p95 demo-api', 'demo-api', '', 'latency_p95', 99.9, 720, 500, ?),
  ('slo-demo-availability', 'Disponibilidade demo-api', 'demo-api', '', 'availability', 99.9, 720, 0, ?)
`, now, now)
	return err
}
