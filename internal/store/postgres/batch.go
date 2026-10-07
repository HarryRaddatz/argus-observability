package postgres

import (
	"context"
	"time"
)

type txKey struct{}

func withTx(ctx context.Context, tx *boundTx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func txFrom(ctx context.Context) *boundTx {
	tx, _ := ctx.Value(txKey{}).(*boundTx)
	return tx
}

// ApplyBatch claims batchID and runs fn in that same transaction.
// A second caller, including a concurrent one, gets applied=false and fn does not run.
// An empty batchID just runs fn.
func (s *Postgres) ApplyBatch(ctx context.Context, batchID string, fn func(context.Context) error) (bool, error) {
	if batchID == "" {
		if err := fn(ctx); err != nil {
			return false, err
		}
		return true, nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx, `
INSERT INTO ingest_batches (batch_id, received_at) VALUES (?, ?)
ON CONFLICT (batch_id) DO NOTHING`, batchID, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if n == 0 {
		return false, nil
	}
	if err := fn(withTx(ctx, tx)); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
