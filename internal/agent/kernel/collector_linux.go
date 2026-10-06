//go:build linux

package kernel

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	bpf "github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

//go:embed bpf/l4.bpf.o
var programObject []byte

// ErrUnsupported means the kernel or the process cannot run the collector.
var ErrUnsupported = errors.New("kernel collector unavailable")

const (
	connectionProgram  = "argus_inet_sock_set_state"
	retransmitProgram  = "argus_tcp_retransmit_skb"
	statSlotEvents     = 0
	statSlotSkipped    = 1
	statSlotMapErrors  = 2
)

type flowKey struct {
	Source    [16]byte
	Dest      [16]byte
	SourcePort uint16
	DestPort  uint16
	Family    uint8
	Direction uint8
}

type flowValue struct {
	Count  uint64
	LastNS uint64
}

// Collector owns the loaded programs and the kernel maps behind them.
type Collector struct {
	collection *bpf.Collection
	links      []link.Link
	flows      *bpf.Map
	stats      *bpf.Map
	delta      *Delta
	degraded   []string
}

// Stats reports collector health so a silent kernel side is visible in the panel.
type Stats struct {
	Events    uint64
	Skipped   uint64
	MapErrors uint64
}

// Load attaches the tracepoints. The caller is expected to fall back to
// log-derived signals when this returns an error.
func Load() (*Collector, error) {
	if err := rlimit.RemoveMemlock(); err != nil && !errors.Is(err, os.ErrPermission) {
		return nil, fmt.Errorf("%w: lock memory: %v", ErrUnsupported, err)
	}

	spec, err := bpf.LoadCollectionSpecFromReader(bytes.NewReader(programObject))
	if err != nil {
		return nil, fmt.Errorf("%w: load object: %v", ErrUnsupported, err)
	}

	collection, degraded, err := loadCollection(spec)
	if err != nil {
		return nil, err
	}

	c := &Collector{
		collection: collection,
		flows:      collection.Maps["flows"],
		stats:      collection.Maps["stats"],
		delta:      NewDelta(),
		degraded:   degraded,
	}
	if c.flows == nil || c.stats == nil {
		collection.Close()
		return nil, fmt.Errorf("%w: maps missing from object", ErrUnsupported)
	}

	attachments := []struct {
		program string
		group   string
		name    string
	}{
		{connectionProgram, "sock", "inet_sock_set_state"},
		{retransmitProgram, "tcp", "tcp_retransmit_skb"},
	}
	for _, a := range attachments {
		program := collection.Programs[a.program]
		if program == nil {
			c.degraded = append(c.degraded, a.name+": program not loaded")
			continue
		}
		handle, err := link.Tracepoint(a.group, a.name, program, nil)
		if err != nil {
			c.degraded = append(c.degraded, fmt.Sprintf("%s: %v", a.name, err))
			continue
		}
		c.links = append(c.links, handle)
	}
	if len(c.links) == 0 {
		c.Close()
		return nil, fmt.Errorf("%w: no tracepoint could be attached (%s)", ErrUnsupported, strings.Join(c.degraded, "; "))
	}
	return c, nil
}

// loadCollection retries without the retransmission program, whose tracepoint
// struct was renamed across kernel versions.
func loadCollection(spec *bpf.CollectionSpec) (*bpf.Collection, []string, error) {
	collection, err := bpf.NewCollection(spec)
	if err == nil {
		return collection, nil, nil
	}
	first := err

	delete(spec.Programs, retransmitProgram)
	collection, err = bpf.NewCollection(spec)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrUnsupported, first)
	}
	return collection, []string{fmt.Sprintf("%s: %v", retransmitProgram, first)}, nil
}

// Degraded lists programs that could not be loaded or attached.
func (c *Collector) Degraded() []string {
	return c.degraded
}

// Close detaches everything. Programs are unloaded when the agent exits too,
// but an explicit close keeps restarts clean.
func (c *Collector) Close() error {
	for _, handle := range c.links {
		_ = handle.Close()
	}
	c.links = nil
	if c.collection != nil {
		c.collection.Close()
		c.collection = nil
	}
	return nil
}

// Flows returns connection counts observed since the previous call.
func (c *Collector) Flows() ([]Flow, error) {
	var (
		key   flowKey
		value flowValue
		out   []Flow
	)
	iter := c.flows.Iterate()
	for iter.Next(&key, &value) {
		out = append(out, Flow{
			Source:     address(key.Source, key.Family),
			Dest:       address(key.Dest, key.Family),
			SourcePort: key.SourcePort,
			DestPort:   key.DestPort,
			Direction:  Direction(key.Direction),
			Count:      value.Count,
		})
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return c.delta.Apply(out), nil
}

// Stats reads the kernel side counters.
func (c *Collector) Stats() Stats {
	return Stats{
		Events:    c.statSlot(statSlotEvents),
		Skipped:   c.statSlot(statSlotSkipped),
		MapErrors: c.statSlot(statSlotMapErrors),
	}
}

func (c *Collector) statSlot(slot uint32) uint64 {
	var value uint64
	if err := c.stats.Lookup(&slot, &value); err != nil {
		return 0
	}
	return value
}

func address(raw [16]byte, family uint8) net.IP {
	if family == 4 {
		return net.IPv4(raw[0], raw[1], raw[2], raw[3])
	}
	ip := make(net.IP, net.IPv6len)
	copy(ip, raw[:])
	return ip
}
