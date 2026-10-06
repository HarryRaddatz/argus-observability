package docker

import (
	"context"
	"net"
	"strings"

	"github.com/HarryRaddatz/argus-observability/internal/agent/kernel"
)

// WorkloadIndex answers which local container owns an address. The kernel
// collector sees addresses only, so this is what turns a flow into a service.
type WorkloadIndex struct {
	byAddress map[string]kernel.Workload
}

// Workload implements kernel.Resolver.
func (w WorkloadIndex) Workload(ip net.IP) (kernel.Workload, bool) {
	if ip == nil || w.byAddress == nil {
		return kernel.Workload{}, false
	}
	workload, ok := w.byAddress[ip.String()]
	return workload, ok
}

// WorkloadIndex builds the address lookup from the container inventory.
func (c *Collector) WorkloadIndex(ctx context.Context) (WorkloadIndex, error) {
	containers, err := c.listContainers(ctx, false)
	if err != nil {
		return WorkloadIndex{}, err
	}
	prefix := nameFilter()
	index := WorkloadIndex{byAddress: map[string]kernel.Workload{}}
	for _, ctr := range containers {
		name := ctr.primaryName()
		if prefix != "" && !strings.HasPrefix(name, prefix) {
			continue
		}
		entityUID, labels := c.entityFor(name, ctr.Labels)
		workload := kernel.Workload{
			EntityUID: entityUID,
			Service:   labels["service"],
			Container: name,
		}
		for _, network := range ctr.NetworkSettings.Networks {
			for _, address := range []string{network.IPAddress, network.GlobalIPv6Address} {
				if address == "" {
					continue
				}
				index.byAddress[address] = workload
			}
		}
	}
	return index, nil
}
