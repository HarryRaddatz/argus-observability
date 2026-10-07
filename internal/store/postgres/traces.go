package postgres

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/insights"
	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/HarryRaddatz/argus-observability/internal/slo"
)

func (s *Postgres) WriteTraceSpans(ctx context.Context, spans []model.TraceSpan) error {
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

func (s *Postgres) GetTraceSpans(ctx context.Context, traceID string) ([]model.TraceSpan, error) {
	norm := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(traceID), "-", ""))
	if norm == "" {
		return nil, nil
	}
	rows, err := s.rdb.QueryContext(ctx, `
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

// traceListScanLimit caps the rows read per source so a wide window stays cheap on large databases.
const traceListScanLimit = 5000

// ListTraces returns recent traces from OTLP spans and from log lines that carry a trace id.
// A trace present in both sources is reported once, from its OTLP spans.
func (s *Postgres) ListTraces(ctx context.Context, filter model.TraceListFilter) (model.TracePage, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	since := filter.Since.UTC().Format(time.RFC3339Nano)

	byKey := map[string]*model.TraceSummary{}
	services := map[string]map[string]struct{}{}
	hasRoot := map[string]bool{}
	addService := func(key, svc string) {
		if svc == "" {
			return
		}
		if services[key] == nil {
			services[key] = map[string]struct{}{}
		}
		services[key][svc] = struct{}{}
	}

	spanRows, err := s.rdb.QueryContext(ctx, `
SELECT trace_id, parent_span_id, name, service, container, start_ts, end_ts, status
FROM trace_spans WHERE start_ts >= ?
ORDER BY start_ts DESC LIMIT ?`, since, traceListScanLimit)
	if err != nil {
		return model.TracePage{}, err
	}
	spanScanned := 0
	for spanRows.Next() {
		spanScanned++
		var traceID, parent, name, service, container, startStr, endStr, status string
		if err := spanRows.Scan(&traceID, &parent, &name, &service, &container, &startStr, &endStr, &status); err != nil {
			spanRows.Close()
			return model.TracePage{}, err
		}
		start, _ := time.Parse(time.RFC3339Nano, startStr)
		end, _ := time.Parse(time.RFC3339Nano, endStr)
		key := traceKey(traceID)
		sum := byKey[key]
		if sum == nil {
			sum = &model.TraceSummary{TraceID: traceID, Source: "otlp", StartTS: start, EndTS: end}
			byKey[key] = sum
		}
		sum.SpanCount++
		if start.Before(sum.StartTS) {
			sum.StartTS = start
		}
		if end.After(sum.EndTS) {
			sum.EndTS = end
		}
		if status == "error" {
			sum.Error = true
		}
		if parent == "" && !hasRoot[key] {
			sum.Name, sum.Service, sum.Container = name, service, container
			hasRoot[key] = true
		} else if !hasRoot[key] {
			sum.Name, sum.Service, sum.Container = name, service, container
		}
		addService(key, service)
		addService(key, insights.InferServiceFromContainer(container))
	}
	if err := spanRows.Err(); err != nil {
		spanRows.Close()
		return model.TracePage{}, err
	}
	spanRows.Close()

	logRows, err := s.rdb.QueryContext(ctx, `
SELECT ts, message, level, entity_uid, labels_json,
  COALESCE(fields_json::json->>'trace_id', fields_json::json->>'traceId', '')
FROM log_entries
WHERE ts >= ? AND (fields_json LIKE '%"trace_id"%' OR fields_json LIKE '%"traceId"%')
ORDER BY ts DESC LIMIT ?`, since, traceListScanLimit)
	if err != nil {
		return model.TracePage{}, err
	}
	logScanned := 0
	defer logRows.Close()
	for logRows.Next() {
		logScanned++
		var tsStr, message, level, entityUID, labelsJSON, traceID string
		if err := logRows.Scan(&tsStr, &message, &level, &entityUID, &labelsJSON, &traceID); err != nil {
			return model.TracePage{}, err
		}
		key := traceKey(traceID)
		if key == "" {
			continue
		}
		sum := byKey[key]
		if sum != nil && sum.Source == "otlp" {
			continue
		}
		ts, _ := time.Parse(time.RFC3339Nano, tsStr)
		if sum == nil {
			sum = &model.TraceSummary{TraceID: traceID, Source: "logs", StartTS: ts, EndTS: ts}
			byKey[key] = sum
		}
		var labels model.Labels
		_ = json.Unmarshal([]byte(labelsJSON), &labels)
		container := containerFromEntityUID(entityUID)
		service := labels["service"]
		if service == "" {
			service = insights.InferServiceFromContainer(container)
		}
		sum.SpanCount++
		if ts.Before(sum.StartTS) {
			sum.StartTS = ts
		}
		if ts.After(sum.EndTS) {
			sum.EndTS = ts
		}
		switch strings.ToLower(level) {
		case "error", "fatal", "critical":
			sum.Error = true
		}
		// Rows arrive newest first, so the last write keeps the earliest line as the trace name.
		sum.Name = logTraceName(message)
		sum.Service, sum.Container = service, container
		addService(key, service)
	}
	if err := logRows.Err(); err != nil {
		return model.TracePage{}, err
	}

	out := make([]model.TraceSummary, 0, len(byKey))
	for key, sum := range byKey {
		if filter.Service != "" {
			if _, ok := services[key][filter.Service]; !ok {
				continue
			}
		}
		if sum.EndTS.After(sum.StartTS) {
			sum.DurationMs = float64(sum.EndTS.Sub(sum.StartTS).Microseconds()) / 1000
		}
		out = append(out, *sum)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartTS.After(out[j].StartTS) })
	total := len(out)
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}
	out = out[offset:]
	if len(out) > limit {
		out = out[:limit]
	}
	return model.TracePage{
		Traces:    out,
		Total:     total,
		Truncated: spanScanned == traceListScanLimit || logScanned == traceListScanLimit,
	}, nil
}

func traceKey(traceID string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(traceID), "-", ""))
}

func containerFromEntityUID(entityUID string) string {
	if i := strings.LastIndex(entityUID, ":"); i >= 0 {
		return entityUID[i+1:]
	}
	return entityUID
}

// logTraceName picks a readable operation name from a log line, preferring the route of structured JSON logs.
func logTraceName(message string) string {
	msg := strings.TrimSpace(message)
	if strings.HasPrefix(msg, "{") {
		var obj map[string]any
		if json.Unmarshal([]byte(msg), &obj) == nil {
			method, _ := obj["method"].(string)
			for _, k := range []string{"route", "path", "url", "name", "msg", "message", "event"} {
				v, ok := obj[k].(string)
				if !ok || v == "" {
					continue
				}
				if method != "" && (k == "route" || k == "path" || k == "url") {
					v = method + " " + v
				}
				return truncateRunes(v, 80)
			}
		}
	}
	return truncateRunes(msg, 80)
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func (s *Postgres) ListSLOs(ctx context.Context) ([]model.SLODefinition, error) {
	rows, err := s.rdb.QueryContext(ctx, `
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

func (s *Postgres) GetSLO(ctx context.Context, id string) (model.SLODefinition, error) {
	row := s.rdb.QueryRowContext(ctx, `
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

func (s *Postgres) EvaluateSLO(ctx context.Context, def model.SLODefinition, at time.Time) (model.SLOStatus, error) {
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

func (s *Postgres) httpMetricsForService(ctx context.Context, service string, since time.Time) ([]float64, int, int, error) {
	rows, err := s.rdb.QueryContext(ctx, `
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

func (s *Postgres) seedDefaultSLOs() error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.Exec(`
INSERT INTO slos (id, name, service, group_id, sli_metric, target, window_hours, latency_threshold_ms, created_at)
VALUES
  ('slo-demo-latency', 'Latência p95 demo-api', 'demo-api', '', 'latency_p95', 99.9, 720, 500, ?),
  ('slo-demo-availability', 'Disponibilidade demo-api', 'demo-api', '', 'availability', 99.9, 720, 0, ?)
ON CONFLICT (id) DO NOTHING
`, now, now)
	return err
}
