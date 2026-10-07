// Package factory opens the store selected by configuration.
//
// The hub keeps a single store.Store. Each driver, relational or document,
// implements that interface and is chosen with ARGUS_STORE_DRIVER. Postgres is
// the built-in relational driver. A document driver registers itself with
// Register and is selected by name.
package factory

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/HarryRaddatz/argus-observability/internal/store"
	"github.com/HarryRaddatz/argus-observability/internal/store/postgres"
)

// Driver names.
const (
	DriverPostgres = "postgres"
)

type opener func(ctx context.Context, dsn string) (store.Store, error)

var (
	mu      sync.RWMutex
	drivers = map[string]opener{}
)

// Register adds a driver. The name is matched case-insensitively against
// ARGUS_STORE_DRIVER. Document stores use the same hook as relational ones.
func Register(name string, open opener) {
	mu.Lock()
	defer mu.Unlock()
	drivers[strings.ToLower(strings.TrimSpace(name))] = open
}

// Open connects the configured driver. An empty driver means postgres.
func Open(ctx context.Context, driver, dsn string) (store.Store, error) {
	name := strings.ToLower(strings.TrimSpace(driver))
	if name == "" {
		name = DriverPostgres
	}
	if name == DriverPostgres {
		return postgres.Open(dsn)
	}
	mu.RLock()
	open, ok := drivers[name]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("store driver %q is not registered", driver)
	}
	return open(ctx, dsn)
}
