package postgres

import (
	"os"
	"testing"

	"github.com/HarryRaddatz/argus-observability/internal/store"
)

func openTest(t *testing.T) store.Store {
	t.Helper()
	dsn := os.Getenv("ARGUS_STORE_DSN")
	if dsn == "" {
		t.Skip("ARGUS_STORE_DSN is not set")
	}
	st, err := OpenIsolated(dsn, SchemaName(t.Name()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}
