package sqlite

import (
	"context"
	"encoding/json"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func (s *SQLite) UpsertFleetStatus(ctx context.Context, rows []model.ContainerFleetStatus) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM container_fleet`); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO container_fleet (entity_uid, container, service, state, health, restart_count, exit_code, oom_killed, status_text, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, r := range rows {
		oom := 0
		if r.OOMKilled {
			oom = 1
		}
		if _, err := stmt.ExecContext(ctx,
			r.EntityUID, r.Container, r.Service, r.State, r.Health,
			r.RestartCount, r.ExitCode, oom, r.StatusText,
			r.UpdatedAt.UTC().Format(time.RFC3339Nano),
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLite) GetFleetStatus(ctx context.Context) ([]model.ContainerFleetStatus, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT entity_uid, container, service, state, health, restart_count, exit_code, oom_killed, status_text, updated_at
FROM container_fleet ORDER BY container ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.ContainerFleetStatus
	for rows.Next() {
		var r model.ContainerFleetStatus
		var tsStr string
		var oom int
		if err := rows.Scan(
			&r.EntityUID, &r.Container, &r.Service, &r.State, &r.Health,
			&r.RestartCount, &r.ExitCode, &oom, &r.StatusText, &tsStr,
		); err != nil {
			return nil, err
		}
		r.OOMKilled = oom == 1
		r.UpdatedAt, _ = time.Parse(time.RFC3339Nano, tsStr)
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *SQLite) CountFleetEvents(ctx context.Context, since time.Time) (model.FleetEventStats, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT type, payload_json FROM events WHERE ts >= ?
`, since.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return model.FleetEventStats{}, err
	}
	defer rows.Close()
	var stats model.FleetEventStats
	for rows.Next() {
		var typ, payloadJSON string
		if err := rows.Scan(&typ, &payloadJSON); err != nil {
			return model.FleetEventStats{}, err
		}
		switch typ {
		case "container.restart":
			stats.Restarts24h++
		case "container.oom":
			stats.OOM24h++
		case "agent.disconnect":
			stats.Disconnect24h++
		case "container.die":
			var payload map[string]any
			_ = json.Unmarshal([]byte(payloadJSON), &payload)
			if code, ok := payload["exitCode"].(string); ok && code != "0" && code != "" {
				stats.Failures24h++
			} else if code, ok := payload["exitCode"].(float64); ok && code != 0 {
				stats.Failures24h++
			}
		}
	}
	return stats, rows.Err()
}
