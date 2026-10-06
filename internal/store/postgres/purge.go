package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

const purgeBatchSize = 2000

func (s *Postgres) Purge(ctx context.Context, logsBefore, metricsBefore, eventsBefore time.Time) (model.PurgeResult, error) {
	start := time.Now()
	out := model.PurgeResult{}

	var err error
	out.LogsDeleted, err = s.purgeBefore(ctx, "log_entries", "ts", logsBefore)
	if err != nil {
		return out, err
	}
	out.MetricsDeleted, err = s.purgeBefore(ctx, "metric_points", "ts", metricsBefore)
	if err != nil {
		return out, err
	}
	out.EventsDeleted, err = s.purgeBefore(ctx, "events", "ts", eventsBefore)
	if err != nil {
		return out, err
	}
	if _, err = s.purgeBefore(ctx, "trace_spans", "end_ts", logsBefore); err != nil {
		return out, err
	}
	if _, err = s.purgeBefore(ctx, "log_patterns", "last_seen", logsBefore); err != nil {
		return out, err
	}
	if _, err = s.purgeBefore(ctx, "topology_edges", "last_seen", metricsBefore); err != nil {
		return out, err
	}
	if _, err = s.purgeBefore(ctx, "topology_links", "last_seen", metricsBefore); err != nil {
		return out, err
	}

	out.Duration = time.Since(start)
	if ctx.Err() != nil {
		out.Truncated = true
		return out, ctx.Err()
	}
	return out, nil
}

func (s *Postgres) purgeBefore(ctx context.Context, table, tsColumn string, before time.Time) (int64, error) {
	cutoff := before.UTC().Format(time.RFC3339Nano)
	if table == "topology_edges" || table == "topology_links" {
		res, err := s.db.ExecContext(ctx,
			fmt.Sprintf(`DELETE FROM %s WHERE %s < ?`, table, tsColumn), cutoff)
		if err != nil {
			return 0, err
		}
		return res.RowsAffected()
	}
	query := fmt.Sprintf(
		`DELETE FROM %s WHERE id IN (SELECT id FROM %s WHERE %s < ? ORDER BY id LIMIT ?)`,
		table, table, tsColumn,
	)
	var total int64
	for {
		if err := ctx.Err(); err != nil {
			return total, err
		}
		res, err := s.db.ExecContext(ctx, query, cutoff, purgeBatchSize)
		if err != nil {
			return total, err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return total, err
		}
		total += n
		if n < purgeBatchSize {
			return total, nil
		}
	}
}
