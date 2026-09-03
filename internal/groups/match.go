package groups

import (
	"github.com/HarryRaddatz/argus-observability/internal/model"
)

// Matches returns true if the workload belongs to the group selector.
func Matches(w model.WorkloadSnapshot, g model.WorkloadGroup) bool {
	switch g.Kind {
	case model.GroupKindStack:
		return labelEquals(w, "stack", g.LabelValue)
	case model.GroupKindService:
		return labelEquals(w, "service", g.LabelValue)
	case model.GroupKindCustom:
		for _, c := range g.Containers {
			if c == w.Container {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func labelEquals(w model.WorkloadSnapshot, key, value string) bool {
	if value == "" {
		return false
	}
	if w.Labels != nil && w.Labels[key] == value {
		return true
	}
	switch key {
	case "stack":
		return w.Stack == value
	case "service":
		return w.Service == value
	}
	return false
}

// FilterWorkloads returns workloads that match the group.
func FilterWorkloads(all []model.WorkloadSnapshot, g model.WorkloadGroup) []model.WorkloadSnapshot {
	var out []model.WorkloadSnapshot
	for _, w := range all {
		if Matches(w, g) {
			out = append(out, w)
		}
	}
	return out
}

// Discover builds suggested groups from current workloads (stack + service).
func Discover(workloads []model.WorkloadSnapshot) []model.WorkloadGroup {
	stackCounts := map[string]int{}
	serviceCounts := map[string]int{}
	for _, w := range workloads {
		stack := w.Stack
		if stack == "" && w.Labels != nil {
			stack = w.Labels["stack"]
		}
		if stack != "" {
			stackCounts[stack]++
		}
		svc := w.Service
		if svc == "" && w.Labels != nil {
			svc = w.Labels["service"]
		}
		if svc != "" {
			serviceCounts[svc]++
		}
	}
	var out []model.WorkloadGroup
	for stack, n := range stackCounts {
		if n == 0 {
			continue
		}
		out = append(out, model.WorkloadGroup{
			ID:          "discover-stack-" + stack,
			Name:        "Stack " + stack,
			Kind:       model.GroupKindStack,
			LabelKey:   "stack",
			LabelValue: stack,
			MemberCount: n,
		})
	}
	for svc, n := range serviceCounts {
		if n == 0 {
			continue
		}
		out = append(out, model.WorkloadGroup{
			ID:         "discover-service-" + svc,
			Name:       "Serviço " + svc,
			Kind:       model.GroupKindService,
			LabelKey:   "service",
			LabelValue: svc,
			MemberCount: n,
		})
	}
	return out
}

// Summarize aggregates metrics for group members.
func Summarize(g model.WorkloadGroup, members []model.WorkloadSnapshot) model.WorkloadGroupSummary {
	var cpuSum, memPctSum, memSum float64
	var memPctN int
	for _, m := range members {
		cpuSum += m.CPUUsage
		memSum += m.MemoryUsage
		if m.MemoryLimit > 0 {
			memPctSum += (m.MemoryUsage / m.MemoryLimit) * 100
			memPctN++
		}
	}
	n := len(members)
	summary := model.WorkloadGroupSummary{
		Group:       g,
		MemberCount: n,
		Members:     members,
	}
	if n > 0 {
		summary.AvgCPU = cpuSum / float64(n)
		summary.TotalMemory = memSum
	}
	if memPctN > 0 {
		summary.AvgMemoryPct = memPctSum / float64(memPctN)
	}
	return summary
}
