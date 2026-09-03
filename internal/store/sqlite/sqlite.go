package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/insights"
	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/HarryRaddatz/argus-observability/internal/store"

	_ "modernc.org/sqlite"
)

type SQLite struct {
	db *sql.DB
}

func Open(path string) (store.Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &SQLite{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *SQLite) Close() error {
	return s.db.Close()
}

func (s *SQLite) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS agents (
  agent_id TEXT PRIMARY KEY,
  host_id TEXT NOT NULL,
  runtime TEXT NOT NULL,
  labels_json TEXT NOT NULL DEFAULT '{}',
  last_seen TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS metric_points (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  metric_name TEXT NOT NULL,
  ts TEXT NOT NULL,
  value REAL NOT NULL,
  entity_uid TEXT NOT NULL,
  labels_json TEXT NOT NULL DEFAULT '{}'
);
CREATE INDEX IF NOT EXISTS idx_metric_name_ts ON metric_points(metric_name, ts);
CREATE TABLE IF NOT EXISTS log_entries (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ts TEXT NOT NULL,
  message TEXT NOT NULL,
  level TEXT NOT NULL DEFAULT 'info',
  entity_uid TEXT NOT NULL,
  labels_json TEXT NOT NULL DEFAULT '{}',
  fields_json TEXT NOT NULL DEFAULT '{}'
);
CREATE INDEX IF NOT EXISTS idx_log_entity_ts ON log_entries(entity_uid, ts);
CREATE INDEX IF NOT EXISTS idx_log_ts ON log_entries(ts);
CREATE TABLE IF NOT EXISTS events (
  id TEXT PRIMARY KEY,
  type TEXT NOT NULL,
  ts TEXT NOT NULL,
  severity TEXT NOT NULL,
  source TEXT NOT NULL,
  entity_uid TEXT NOT NULL,
  labels_json TEXT NOT NULL DEFAULT '{}',
  payload_json TEXT NOT NULL DEFAULT '{}'
);
CREATE INDEX IF NOT EXISTS idx_events_entity_ts ON events(entity_uid, ts);
CREATE INDEX IF NOT EXISTS idx_events_ts ON events(ts);
CREATE TABLE IF NOT EXISTS container_fleet (
  entity_uid TEXT PRIMARY KEY,
  container TEXT NOT NULL,
  service TEXT NOT NULL DEFAULT '',
  state TEXT NOT NULL,
  health TEXT NOT NULL DEFAULT '',
  restart_count INTEGER NOT NULL DEFAULT 0,
  exit_code INTEGER NOT NULL DEFAULT 0,
  oom_killed INTEGER NOT NULL DEFAULT 0,
  status_text TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_fleet_service ON container_fleet(service);
CREATE TABLE IF NOT EXISTS workload_groups (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  kind TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  label_key TEXT NOT NULL DEFAULT '',
  label_value TEXT NOT NULL DEFAULT '',
  containers_json TEXT NOT NULL DEFAULT '[]',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
`
	_, err := s.db.Exec(schema)
	if err != nil {
		return err
	}
	return s.seedDefaultGroups()
}

func (s *SQLite) seedDefaultGroups() error {
	return nil
}

func (s *SQLite) UpsertAgent(ctx context.Context, reg model.AgentRegistration) error {
	labels, err := json.Marshal(reg.Labels)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO agents (agent_id, host_id, runtime, labels_json, last_seen)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(agent_id) DO UPDATE SET
  host_id=excluded.host_id,
  runtime=excluded.runtime,
  labels_json=excluded.labels_json,
  last_seen=excluded.last_seen
`, reg.AgentID, reg.HostID, reg.Runtime, string(labels), reg.LastSeen.UTC().Format(time.RFC3339Nano))
	return err
}

func (s *SQLite) TouchAgent(ctx context.Context, agentID string, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE agents SET last_seen=? WHERE agent_id=?`,
		at.UTC().Format(time.RFC3339Nano), agentID)
	return err
}

func (s *SQLite) GetAgent(ctx context.Context, agentID string) (model.AgentRegistration, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT agent_id, host_id, runtime, labels_json, last_seen FROM agents WHERE agent_id=?
`, agentID)
	var reg model.AgentRegistration
	var labelsJSON, lastSeen string
	if err := row.Scan(&reg.AgentID, &reg.HostID, &reg.Runtime, &labelsJSON, &lastSeen); err != nil {
		return model.AgentRegistration{}, err
	}
	_ = json.Unmarshal([]byte(labelsJSON), &reg.Labels)
	reg.LastSeen, _ = time.Parse(time.RFC3339Nano, lastSeen)
	return reg, nil
}

func (s *SQLite) StaleAgents(ctx context.Context, before time.Time) ([]model.AgentRegistration, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT agent_id, host_id, runtime, labels_json, last_seen FROM agents WHERE last_seen < ?
`, before.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.AgentRegistration
	for rows.Next() {
		var reg model.AgentRegistration
		var labelsJSON, lastSeen string
		if err := rows.Scan(&reg.AgentID, &reg.HostID, &reg.Runtime, &labelsJSON, &lastSeen); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(labelsJSON), &reg.Labels)
		reg.LastSeen, _ = time.Parse(time.RFC3339Nano, lastSeen)
		out = append(out, reg)
	}
	return out, rows.Err()
}

func (s *SQLite) WriteMetrics(ctx context.Context, points []model.MetricPoint) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO metric_points (metric_name, ts, value, entity_uid, labels_json)
VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, p := range points {
		labels, err := json.Marshal(p.Labels)
		if err != nil {
			return err
		}
		if _, err := stmt.ExecContext(ctx, p.MetricName, p.TS.UTC().Format(time.RFC3339Nano), p.Value, p.EntityUID, string(labels)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLite) WriteLogs(ctx context.Context, entries []model.LogEntry) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO log_entries (ts, message, level, entity_uid, labels_json, fields_json)
VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, e := range entries {
		labels, err := json.Marshal(e.Labels)
		if err != nil {
			return err
		}
		fields, err := json.Marshal(e.Fields)
		if err != nil {
			return err
		}
		if _, err := stmt.ExecContext(ctx, e.TS.UTC().Format(time.RFC3339Nano), e.Message, e.Level, e.EntityUID, string(labels), string(fields)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLite) WriteEvents(ctx context.Context, events []model.Event) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	stmt, err := tx.PrepareContext(ctx, `
INSERT OR REPLACE INTO events (id, type, ts, severity, source, entity_uid, labels_json, payload_json)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, e := range events {
		labels, err := json.Marshal(e.Labels)
		if err != nil {
			return err
		}
		payload, err := json.Marshal(e.Payload)
		if err != nil {
			return err
		}
		if _, err := stmt.ExecContext(ctx, e.ID, e.Type, e.TS.UTC().Format(time.RFC3339Nano), e.Severity, e.Source, e.EntityUID, string(labels), string(payload)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLite) QueryMetrics(ctx context.Context, metricName string, labels model.Labels, since time.Time) ([]model.SeriesPoint, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT ts, value FROM metric_points
WHERE metric_name=? AND ts >= ?
ORDER BY ts ASC
`, metricName, since.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.SeriesPoint, 0)
	for rows.Next() {
		var tsStr string
		var p model.SeriesPoint
		if err := rows.Scan(&tsStr, &p.Value); err != nil {
			return nil, err
		}
		p.TS, _ = time.Parse(time.RFC3339Nano, tsStr)
		out = append(out, p)
	}
	if labels != nil && len(labels) > 0 {
		_ = labels // label filter in v2; v1 returns all for metric+since
	}
	return out, rows.Err()
}

func (s *SQLite) QueryMetricSeries(ctx context.Context, metricName, container string, since time.Time) ([]model.ContainerSeries, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT ts, value, entity_uid, labels_json FROM metric_points
WHERE metric_name=? AND ts >= ?
ORDER BY ts ASC
`, metricName, since.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type key struct{ container, entity string }
	series := map[key]*model.ContainerSeries{}

	for rows.Next() {
		var tsStr, labelsJSON, entityUID string
		var value float64
		if err := rows.Scan(&tsStr, &value, &entityUID, &labelsJSON); err != nil {
			return nil, err
		}
		var labels model.Labels
		_ = json.Unmarshal([]byte(labelsJSON), &labels)
		name := labels["container"]
		if strings.HasPrefix(metricName, "http.") {
			if svc := labels["service"]; svc != "" {
				name = svc
			}
		}
		if name == "" {
			name = entityUID
		}
		if container != "" && name != container && labels["container"] != container && labels["service"] != container {
			continue
		}
		ts, _ := time.Parse(time.RFC3339Nano, tsStr)
		k := key{container: name, entity: entityUID}
		if series[k] == nil {
			series[k] = &model.ContainerSeries{Container: name, EntityUID: entityUID, Points: make([]model.SeriesPoint, 0)}
		}
		series[k].Points = append(series[k].Points, model.SeriesPoint{TS: ts, Value: value})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]model.ContainerSeries, 0, len(series))
	for _, v := range series {
		out = append(out, *v)
	}
	return out, nil
}

func (s *SQLite) ListWorkloads(ctx context.Context, since time.Time) ([]model.WorkloadSnapshot, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT ts, metric_name, value, entity_uid, labels_json FROM metric_points
WHERE ts >= ? AND metric_name IN ('cpu.usage', 'memory.usage', 'memory.limit')
ORDER BY ts DESC
`, since.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type key string
	latest := map[key]*model.WorkloadSnapshot{}
	seenMetric := map[key]map[string]bool{}

	for rows.Next() {
		var tsStr, metricName, entityUID, labelsJSON string
		var value float64
		if err := rows.Scan(&tsStr, &metricName, &value, &entityUID, &labelsJSON); err != nil {
			return nil, err
		}
		var labels model.Labels
		_ = json.Unmarshal([]byte(labelsJSON), &labels)
		name := labels["container"]
		if name == "" {
			continue
		}
		k := key(name)
		if seenMetric[k] == nil {
			seenMetric[k] = map[string]bool{}
		}
		if seenMetric[k][metricName] {
			continue
		}
		seenMetric[k][metricName] = true
		ts, _ := time.Parse(time.RFC3339Nano, tsStr)
		if latest[k] == nil {
			latest[k] = &model.WorkloadSnapshot{
				Container: name, EntityUID: entityUID, UpdatedAt: ts,
				Labels: labels, Stack: labels["stack"], Service: labels["service"],
			}
		}
		switch metricName {
		case "cpu.usage":
			latest[k].CPUUsage = value
		case "memory.usage":
			latest[k].MemoryUsage = value
		case "memory.limit":
			latest[k].MemoryLimit = value
		}
		if ts.After(latest[k].UpdatedAt) {
			latest[k].UpdatedAt = ts
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]model.WorkloadSnapshot, 0, len(latest))
	for _, w := range latest {
		out = append(out, *w)
	}
	return out, nil
}

func (s *SQLite) ListEvents(ctx context.Context, entityUID string, since time.Time, limit int) ([]model.Event, error) {
	if limit <= 0 {
		limit = 100
	}
	q := `SELECT id, type, ts, severity, source, entity_uid, labels_json, payload_json FROM events WHERE ts >= ?`
	args := []any{since.UTC().Format(time.RFC3339Nano)}
	if entityUID != "" {
		q += ` AND entity_uid=?`
		args = append(args, entityUID)
	}
	q += ` ORDER BY ts DESC LIMIT ?`
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func (s *SQLite) SearchLogs(ctx context.Context, filter model.LogSearchFilter) ([]model.LogEntry, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 200
	}
	q := `SELECT ts, message, level, entity_uid, labels_json, fields_json FROM log_entries WHERE ts >= ?`
	args := []any{filter.Since.UTC().Format(time.RFC3339Nano)}

	if filter.Query != "" {
		q += ` AND message LIKE ?`
		args = append(args, fmt.Sprintf("%%%s%%", filter.Query))
	}
	if filter.EntityUID != "" {
		q += ` AND entity_uid=?`
		args = append(args, filter.EntityUID)
	}
	if filter.Container != "" {
		q += ` AND entity_uid LIKE ?`
		args = append(args, "%:"+filter.Container)
	}
	if filter.Level != "" && filter.Level != "all" {
		q += ` AND level=?`
		args = append(args, filter.Level)
	}
	if filter.Topic != "" && filter.Topic != "all" {
		q += ` AND fields_json LIKE ?`
		args = append(args, fmt.Sprintf(`%%"topics":%%"%s"%%`, filter.Topic))
	}
	if filter.TraceID != "" {
		patterns := insights.TraceSearchPatterns(filter.TraceID)
		if len(patterns) > 0 {
			q += ` AND (`
			for i, p := range patterns {
				if i > 0 {
					q += ` OR `
				}
				q += `(fields_json LIKE ? OR message LIKE ?)`
				args = append(args, p, p)
			}
			q += `)`
		}
	}
	if len(filter.Containers) > 0 {
		q += ` AND (`
		for i, c := range filter.Containers {
			if i > 0 {
				q += ` OR `
			}
			q += `entity_uid LIKE ?`
			args = append(args, "%:"+c)
		}
		q += `)`
	}

	q += ` ORDER BY ts DESC LIMIT ?`
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.LogEntry
	out = make([]model.LogEntry, 0)
	for rows.Next() {
		var e model.LogEntry
		var tsStr, labelsJSON, fieldsJSON string
		if err := rows.Scan(&tsStr, &e.Message, &e.Level, &e.EntityUID, &labelsJSON, &fieldsJSON); err != nil {
			return nil, err
		}
		e.TS, _ = time.Parse(time.RFC3339Nano, tsStr)
		_ = json.Unmarshal([]byte(labelsJSON), &e.Labels)
		_ = json.Unmarshal([]byte(fieldsJSON), &e.Fields)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *SQLite) CountLogTopics(ctx context.Context, since time.Time) ([]model.LogTopicCount, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT entity_uid, fields_json FROM log_entries WHERE ts >= ?
`, since.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type key struct {
		container string
		entityUID string
		topic     string
	}
	counts := map[key]int{}
	for rows.Next() {
		var entityUID, fieldsJSON string
		if err := rows.Scan(&entityUID, &fieldsJSON); err != nil {
			return nil, err
		}
		var fields map[string]any
		_ = json.Unmarshal([]byte(fieldsJSON), &fields)
		rawTopics, _ := fields["topics"].([]any)
		container := entityUID
		if parts := strings.Split(entityUID, ":"); len(parts) > 0 {
			container = parts[len(parts)-1]
		}
		if len(rawTopics) == 0 {
			continue
		}
		for _, t := range rawTopics {
			topic, _ := t.(string)
			if topic == "" || topic == "general" {
				continue
			}
			k := key{container: container, entityUID: entityUID, topic: topic}
			counts[k]++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]model.LogTopicCount, 0, len(counts))
	for k, n := range counts {
		out = append(out, model.LogTopicCount{
			Container: k.container,
			EntityUID: k.entityUID,
			Topic:     k.topic,
			Count:     n,
		})
	}
	return out, nil
}

func scanEvents(rows *sql.Rows) ([]model.Event, error) {
	out := make([]model.Event, 0)
	for rows.Next() {
		var e model.Event
		var tsStr, labelsJSON, payloadJSON string
		if err := rows.Scan(&e.ID, &e.Type, &tsStr, &e.Severity, &e.Source, &e.EntityUID, &labelsJSON, &payloadJSON); err != nil {
			return nil, err
		}
		e.TS, _ = time.Parse(time.RFC3339Nano, tsStr)
		_ = json.Unmarshal([]byte(labelsJSON), &e.Labels)
		_ = json.Unmarshal([]byte(payloadJSON), &e.Payload)
		out = append(out, e)
	}
	return out, rows.Err()
}
