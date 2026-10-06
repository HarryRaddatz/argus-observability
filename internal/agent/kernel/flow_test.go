package kernel

import (
	"net"
	"testing"
	"time"
)

type staticResolver map[string]Workload

func (r staticResolver) Workload(ip net.IP) (Workload, bool) {
	if ip == nil {
		return Workload{}, false
	}
	workload, ok := r[ip.String()]
	return workload, ok
}

func TestBuildSample(t *testing.T) {
	resolver := staticResolver{
		"10.0.0.2": {EntityUID: "docker:host:gateway", Service: "gateway", Container: "stack-gateway-1"},
		"10.0.0.3": {EntityUID: "docker:host:api", Service: "api", Container: "stack-api-1"},
	}
	now := time.Now().UTC()

	flows := []Flow{
		// Both sides of the same connection report it; only the outbound count survives.
		{Source: net.ParseIP("10.0.0.2"), Dest: net.ParseIP("10.0.0.3"), DestPort: 8080, Direction: DirectionOut, Count: 5},
		{Source: net.ParseIP("10.0.0.3"), Dest: net.ParseIP("10.0.0.2"), SourcePort: 8080, Direction: DirectionIn, Count: 5},
		// External dependency keeps the raw address.
		{Source: net.ParseIP("10.0.0.3"), Dest: net.ParseIP("192.0.2.10"), DestPort: 27017, Direction: DirectionOut, Count: 2},
		// A self connection carries no dependency.
		{Source: net.ParseIP("10.0.0.2"), Dest: net.ParseIP("10.0.0.2"), DestPort: 4369, Direction: DirectionOut, Count: 9},
		{Source: net.ParseIP("10.0.0.2"), Dest: net.ParseIP("10.0.0.3"), DestPort: 8080, Direction: DirectionRetransmit, Count: 3},
		{Source: net.ParseIP("10.0.0.2"), Dest: net.ParseIP("10.0.0.3"), DestPort: 8080, Direction: DirectionOut, Count: 0},
	}

	sample := BuildSample(flows, resolver, "host", now)

	if len(sample.Links) != 2 {
		t.Fatalf("links: got %d, want 2 (%+v)", len(sample.Links), sample.Links)
	}
	first := sample.Links[0]
	if first.Source != "gateway" || first.Target != "api" || first.Port != 8080 || first.Kind != "http" || first.Count != 5 {
		t.Fatalf("internal edge: %+v", first)
	}
	second := sample.Links[1]
	if second.Source != "api" || second.Target != "192.0.2.10" || second.Kind != "mongodb" {
		t.Fatalf("external edge: %+v", second)
	}

	metrics := map[string]float64{}
	for _, point := range sample.Metrics {
		metrics[point.MetricName+":"+point.EntityUID] = point.Value
		if point.Labels["source"] != "ebpf" || point.Labels["host"] != "host" {
			t.Fatalf("metric labels: %+v", point.Labels)
		}
	}
	if metrics["net.connections.out:docker:host:gateway"] != 5 || metrics["net.connections.in:docker:host:api"] != 5 {
		t.Fatalf("connection metrics: %+v", metrics)
	}
	if metrics["net.retransmits:docker:host:gateway"] != 3 {
		t.Fatalf("retransmit metric: %+v", metrics)
	}
}

func TestDelta(t *testing.T) {
	delta := NewDelta()
	flow := Flow{Source: net.ParseIP("10.0.0.2"), Dest: net.ParseIP("10.0.0.3"), DestPort: 8080, Count: 10}

	first := delta.Apply([]Flow{flow})
	if len(first) != 1 || first[0].Count != 10 {
		t.Fatalf("first window: %+v", first)
	}

	flow.Count = 14
	second := delta.Apply([]Flow{flow})
	if len(second) != 1 || second[0].Count != 4 {
		t.Fatalf("increment: %+v", second)
	}

	unchanged := delta.Apply([]Flow{flow})
	if len(unchanged) != 0 {
		t.Fatalf("no change should report nothing: %+v", unchanged)
	}

	// An LRU eviction restarts the counter below the previous reading.
	flow.Count = 2
	restarted := delta.Apply([]Flow{flow})
	if len(restarted) != 1 || restarted[0].Count != 2 {
		t.Fatalf("restart after eviction: %+v", restarted)
	}
}

func TestKindForPort(t *testing.T) {
	cases := []struct {
		port uint16
		kind string
	}{
		{6379, "redis"},
		{27017, "mongodb"},
		{8443, "http"},
		{1234, "tcp"},
	}
	for _, c := range cases {
		if got := KindForPort(c.port); got != c.kind {
			t.Errorf("port %d: got %s, want %s", c.port, got, c.kind)
		}
	}
}
