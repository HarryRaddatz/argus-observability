package postgres

import (
	"context"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func (s *Postgres) UpsertFleetStatus(ctx context.Context, rows []model.ContainerFleetStatus) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM container_fleet`); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO container_fleet (entity_uid, container, service, state, health, restart_count, exit_code, oom_killed, status_text, disposition, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
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
			r.RestartCount, r.ExitCode, oom, r.StatusText, r.Disposition,
			r.UpdatedAt.UTC().Format(time.RFC3339Nano),
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Postgres) GetFleetStatus(ctx context.Context) ([]model.ContainerFleetStatus, error) {
	rows, err := s.rdb.QueryContext(ctx, `
SELECT entity_uid, container, service, state, health, restart_count, exit_code, oom_killed, status_text, disposition, updated_at
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
			&r.RestartCount, &r.ExitCode, &oom, &r.StatusText, &r.Disposition, &tsStr,
		); err != nil {
			return nil, err
		}
		r.OOMKilled = oom == 1
		r.UpdatedAt, _ = time.Parse(time.RFC3339Nano, tsStr)
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Postgres) CountFleetEvents(ctx context.Context, since time.Time) (model.FleetEventStats, error) {
	// One pass. FILTER keeps each count on its own rows, and the exit code
	// stays in SQL instead of shipping every payload to the process.
	row := s.rdb.QueryRowContext(ctx, `
SELECT
  count(*) FILTER (WHERE type = 'container.restart'),
  count(*) FILTER (WHERE type = 'container.oom'),
  count(*) FILTER (WHERE type = 'agent.disconnect'),
  count(*) FILTER (WHERE type = 'container.die' AND (
    payload_json::json->>'cause' = 'unexpected'
    OR (
      COALESCE(payload_json::json->>'cause', '') = ''
      AND COALESCE(payload_json::json->>'exitCode', '') <> ''
      AND COALESCE(payload_json::json->>'exitCode', '') NOT IN ('0', '130', '143')
    )
  ))
FROM events
WHERE ts >= ?
`, since.UTC().Format(time.RFC3339Nano))
	var stats model.FleetEventStats
	if err := row.Scan(&stats.Restarts24h, &stats.OOM24h, &stats.Disconnect24h, &stats.Failures24h); err != nil {
		return model.FleetEventStats{}, err
	}
	return stats, nil
}
