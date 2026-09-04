package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

const purgeBatchSize = 2000

func (s *SQLite) Purge(ctx context.Context, logsBefore, metricsBefore, eventsBefore time.Time) (model.PurgeResult, error) {
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

	out.Duration = time.Since(start)
	if ctx.Err() != nil {
		out.Truncated = true
		return out, ctx.Err()
	}
	return out, nil
}

func (s *SQLite) purgeBefore(ctx context.Context, table, tsColumn string, before time.Time) (int64, error) {
	cutoff := before.UTC().Format(time.RFC3339Nano)
	query := fmt.Sprintf(
		`DELETE FROM %s WHERE rowid IN (SELECT rowid FROM %s WHERE %s < ? LIMIT ?)`,
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
