//go:build !linux

package kernel

import "errors"

// ErrUnsupported means the kernel or the process cannot run the collector.
var ErrUnsupported = errors.New("kernel collector requires linux")

// Collector is a no-op outside Linux so the agent still builds everywhere.
type Collector struct{}

// Stats reports collector health so a silent kernel side is visible in the panel.
type Stats struct {
	Events    uint64
	Skipped   uint64
	MapErrors uint64
}

// Load always fails outside Linux.
func Load() (*Collector, error) { return nil, ErrUnsupported }

// Degraded lists programs that could not be loaded or attached.
func (c *Collector) Degraded() []string { return nil }

// Close detaches everything.
func (c *Collector) Close() error { return nil }

// Flows returns connection counts observed since the previous call.
func (c *Collector) Flows() ([]Flow, error) { return nil, ErrUnsupported }

// Stats reads the kernel side counters.
func (c *Collector) Stats() Stats { return Stats{} }
