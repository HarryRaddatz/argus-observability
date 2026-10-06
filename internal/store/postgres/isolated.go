package postgres

import (
	"database/sql"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/HarryRaddatz/argus-observability/internal/store"
)

var schemaName = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

// SchemaName turns an arbitrary label into a safe schema identifier.
func SchemaName(name string) string {
	var b strings.Builder
	b.WriteString("t_")
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	out := b.String()
	if len(out) > 60 {
		out = out[:60]
	}
	return out
}

// OpenIsolated creates a private schema and opens the store inside it.
// Tests use it so they do not share tables.
func OpenIsolated(baseDSN, schema string) (store.Store, error) {
	return openSchema(baseDSN, schema)
}

func openSchema(baseDSN, schema string) (*Postgres, error) {
	if !schemaName.MatchString(schema) {
		return nil, fmt.Errorf("invalid schema %q", schema)
	}
	admin, err := sql.Open("pgx", baseDSN)
	if err != nil {
		return nil, err
	}
	defer admin.Close()
	if err := admin.Ping(); err != nil {
		return nil, err
	}
	if _, err := admin.Exec("DROP SCHEMA IF EXISTS " + schema + " CASCADE"); err != nil {
		return nil, err
	}
	if _, err := admin.Exec("CREATE SCHEMA " + schema); err != nil {
		return nil, err
	}

	u, err := url.Parse(baseDSN)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	opened, err := Open(u.String())
	if err != nil {
		return nil, err
	}
	pg, ok := opened.(*Postgres)
	if !ok {
		_ = opened.Close()
		return nil, fmt.Errorf("unexpected store type")
	}
	return pg, nil
}
