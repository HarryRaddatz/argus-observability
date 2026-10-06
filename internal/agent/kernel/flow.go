// Package kernel collects L4 connection activity straight from kernel tracepoints,
// so topology and connection health do not depend on applications logging them.
package kernel

import (
	"net"
	"sort"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

// Direction mirrors the direction byte written by the kernel program.
type Direction uint8

const (
	// DirectionOut is a connection opened by a local workload.
	DirectionOut Direction = 0
	// DirectionIn is a connection accepted by a local workload.
	DirectionIn Direction = 1
	// DirectionRetransmit is a TCP retransmission, not a new connection.
	DirectionRetransmit Direction = 2
)

// Flow is one aggregated kernel map entry.
type Flow struct {
	Source    net.IP
	Dest      net.IP
	SourcePort uint16
	DestPort  uint16
	Direction Direction
	Count     uint64
}

// ServicePort is the listening side of the connection.
func (f Flow) ServicePort() uint16 {
	if f.Direction == DirectionIn {
		return f.SourcePort
	}
	return f.DestPort
}

// Workload identifies who owns an address on this host.
type Workload struct {
	EntityUID string
	Service   string
	Container string
}

// Resolver maps an address to a local workload. Addresses that belong to
// something else on the network resolve to false.
type Resolver interface {
	Workload(ip net.IP) (Workload, bool)
}

// Sample is one collection window, ready to be shipped to the hub.
type Sample struct {
	Links   []model.TopologyLink
	Metrics []model.MetricPoint
}

type edgeKey struct {
	source string
	target string
	port   uint16
}

type edgeAgg struct {
	link      model.TopologyLink
	fromLocal bool
}

// BuildSample turns raw kernel flows into topology links and per-workload metrics.
//
// A connection between two local containers is reported by both sides, so the
// outbound observation wins and the inbound duplicate is dropped.
func BuildSample(flows []Flow, resolver Resolver, hostID string, ts time.Time) Sample {
	edges := map[edgeKey]*edgeAgg{}
	outbound := map[string]uint64{}
	inbound := map[string]uint64{}
	retransmits := map[string]uint64{}
	owners := map[string]Workload{}

	for _, flow := range flows {
		if flow.Count == 0 {
			continue
		}
		src, srcLocal := resolver.Workload(flow.Source)
		dst, dstLocal := resolver.Workload(flow.Dest)

		if flow.Direction == DirectionRetransmit {
			if srcLocal {
				retransmits[src.EntityUID] += flow.Count
				owners[src.EntityUID] = src
			}
			continue
		}

		// Inbound flows are seen from the server side: source is the listener.
		client, server := src, dst
		clientLocal, serverLocal := srcLocal, dstLocal
		if flow.Direction == DirectionIn {
			client, server = dst, src
			clientLocal, serverLocal = dstLocal, srcLocal
		}

		// A workload talking to itself (epmd, local health checks) carries no dependency.
		self := clientLocal && serverLocal && client.EntityUID == server.EntityUID

		if flow.Direction == DirectionOut && clientLocal && !self {
			outbound[client.EntityUID] += flow.Count
			owners[client.EntityUID] = client
		}
		if flow.Direction == DirectionIn && serverLocal && !self {
			inbound[server.EntityUID] += flow.Count
			owners[server.EntityUID] = server
		}

		sourceName := nodeName(client, clientLocal, flow.clientIP())
		targetName := nodeName(server, serverLocal, flow.serverIP())
		if sourceName == "" || targetName == "" || sourceName == targetName {
			continue
		}

		port := flow.ServicePort()
		key := edgeKey{source: sourceName, target: targetName, port: port}
		agg, ok := edges[key]
		if !ok {
			edges[key] = &edgeAgg{
				link: model.TopologyLink{
					Source: sourceName,
					Target: targetName,
					Kind:   KindForPort(port),
					Port:   port,
					Count:  flow.Count,
					TS:     ts,
				},
				fromLocal: flow.Direction == DirectionOut,
			}
			continue
		}
		// Both directions seen locally: keep the client-side count only.
		if agg.fromLocal && flow.Direction == DirectionIn {
			continue
		}
		if !agg.fromLocal && flow.Direction == DirectionOut {
			agg.link.Count = flow.Count
			agg.fromLocal = true
			continue
		}
		agg.link.Count += flow.Count
	}

	sample := Sample{Links: make([]model.TopologyLink, 0, len(edges))}
	for _, agg := range edges {
		sample.Links = append(sample.Links, agg.link)
	}
	sort.Slice(sample.Links, func(i, j int) bool {
		if sample.Links[i].Count != sample.Links[j].Count {
			return sample.Links[i].Count > sample.Links[j].Count
		}
		return sample.Links[i].Source+sample.Links[i].Target < sample.Links[j].Source+sample.Links[j].Target
	})

	sample.Metrics = append(sample.Metrics, metricsFor(owners, outbound, "net.connections.out", hostID, ts)...)
	sample.Metrics = append(sample.Metrics, metricsFor(owners, inbound, "net.connections.in", hostID, ts)...)
	sample.Metrics = append(sample.Metrics, metricsFor(owners, retransmits, "net.retransmits", hostID, ts)...)
	return sample
}

func (f Flow) clientIP() net.IP {
	if f.Direction == DirectionIn {
		return f.Dest
	}
	return f.Source
}

func (f Flow) serverIP() net.IP {
	if f.Direction == DirectionIn {
		return f.Source
	}
	return f.Dest
}

// nodeName prefers the compose service name, falls back to the container name,
// and finally to the raw address for peers outside this host.
func nodeName(w Workload, local bool, ip net.IP) string {
	if local {
		if w.Service != "" {
			return w.Service
		}
		if w.Container != "" {
			return w.Container
		}
	}
	if ip == nil || ip.IsUnspecified() {
		return ""
	}
	if ip.IsLoopback() {
		return "localhost"
	}
	return ip.String()
}

func metricsFor(owners map[string]Workload, counts map[string]uint64, name, hostID string, ts time.Time) []model.MetricPoint {
	if len(counts) == 0 {
		return nil
	}
	out := make([]model.MetricPoint, 0, len(counts))
	for entityUID, count := range counts {
		owner := owners[entityUID]
		labels := model.Labels{"host": hostID, "runtime": "docker", "source": "ebpf"}
		if owner.Container != "" {
			labels["container"] = owner.Container
		}
		if owner.Service != "" {
			labels["service"] = owner.Service
		}
		out = append(out, model.MetricPoint{
			MetricName: name,
			TS:         ts,
			Value:      float64(count),
			EntityUID:  entityUID,
			Labels:     labels,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].EntityUID < out[j].EntityUID })
	return out
}

// Delta keeps the previous counter reading so each window reports new activity only.
type Delta struct {
	previous map[string]uint64
}

// NewDelta creates an empty counter tracker.
func NewDelta() *Delta {
	return &Delta{previous: map[string]uint64{}}
}

// Apply rewrites cumulative counters into per-window increments.
func (d *Delta) Apply(flows []Flow) []Flow {
	current := make(map[string]uint64, len(flows))
	out := make([]Flow, 0, len(flows))
	for _, flow := range flows {
		key := flowKeyString(flow)
		current[key] = flow.Count
		previous := d.previous[key]
		if flow.Count <= previous {
			// LRU eviction or counter reset: treat the reading as a fresh start.
			if flow.Count == previous {
				continue
			}
		} else {
			flow.Count -= previous
		}
		if flow.Count == 0 {
			continue
		}
		out = append(out, flow)
	}
	d.previous = current
	return out
}

func flowKeyString(f Flow) string {
	buf := make([]byte, 0, 48)
	buf = append(buf, f.Source.To16()...)
	buf = append(buf, f.Dest.To16()...)
	buf = append(buf, byte(f.SourcePort>>8), byte(f.SourcePort), byte(f.DestPort>>8), byte(f.DestPort), byte(f.Direction))
	return string(buf)
}
