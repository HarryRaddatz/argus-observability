package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
)

// boundDB translates the '?' placeholders kept from the previous engine into
// the $1 form Postgres expects. Every connection from the pool goes through it.
type boundDB struct {
	*sql.DB
}

func (b *boundDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return b.DB.ExecContext(ctx, bind(query), args...)
}

func (b *boundDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return b.DB.QueryContext(ctx, bind(query), args...)
}

func (b *boundDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return b.DB.QueryRowContext(ctx, bind(query), args...)
}

func (b *boundDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return b.DB.PrepareContext(ctx, bind(query))
}

func (b *boundDB) Exec(query string, args ...any) (sql.Result, error) {
	return b.DB.Exec(bind(query), args...)
}

func (b *boundDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*boundTx, error) {
	tx, err := b.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &boundTx{Tx: tx}, nil
}

type boundTx struct {
	*sql.Tx
}

func (t *boundTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.Tx.ExecContext(ctx, bind(query), args...)
}

func (t *boundTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return t.Tx.PrepareContext(ctx, bind(query))
}

func bind(query string) string {
	if !strings.Contains(query, "?") {
		return query
	}
	var b strings.Builder
	b.Grow(len(query) + 8)
	n := 0
	for _, r := range query {
		if r == '?' {
			n++
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(n))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
