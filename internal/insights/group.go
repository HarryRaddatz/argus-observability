package insights

import (
	"fmt"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

// GenerateGroup builds aggregated insights for a workload group.
func GenerateGroup(
	g model.WorkloadGroup,
	members []model.WorkloadSnapshot,
	logStats []LogTopicStats,
	fleet []model.ContainerFleetStatus,
) []model.Insight {
	if len(members) == 0 {
		return nil
	}

	memberSet := map[string]struct{}{}
	for _, m := range members {
		memberSet[m.Container] = struct{}{}
	}

	var wlMem []WorkloadMemory
	var pressureCount, criticalMem, hotCPU int
	var cpuSum, memPctSum float64
	var memPctN int

	for _, m := range members {
		pct := 0.0
		if m.MemoryLimit > 0 {
			pct = (m.MemoryUsage / m.MemoryLimit) * 100
			memPctSum += pct
			memPctN++
		}
		cpuSum += m.CPUUsage
		wlMem = append(wlMem, WorkloadMemory{
			Container: m.Container, EntityUID: m.EntityUID,
			CPUUsage: m.CPUUsage, MemoryPct: pct,
			MemoryUsage: m.MemoryUsage, MemoryLimit: m.MemoryLimit,
		})
		if pct >= 90 {
			criticalMem++
			pressureCount++
		} else if pct >= 75 {
			pressureCount++
		}
		if m.CPUUsage >= 80 {
			hotCPU++
			pressureCount++
		}
	}

	var errLines, gcLines int
	for _, s := range logStats {
		if _, ok := memberSet[s.Container]; !ok {
			continue
		}
		switch s.Topic {
		case "error":
			errLines += s.Count
		case "gc":
			gcLines += s.Count
		}
	}

	var out []model.Insight
	groupLabel := g.Name
	evidence := map[string]any{
		"group_id": g.ID, "group_name": g.Name, "member_count": len(members),
	}

	n := float64(len(members))
	if memPctN > 0 {
		evidence["avg_memory_pct"] = memPctSum / float64(memPctN)
	}
	evidence["avg_cpu"] = cpuSum / n

	if criticalMem >= 2 || (criticalMem >= 1 && len(members) <= 3) {
		out = append(out, model.Insight{
			ID: g.ID + ":group_memory_critical", Theme: "group_degradation", Severity: "critical",
			Title:   fmt.Sprintf("Critical memory in group — %s", groupLabel),
			Summary: fmt.Sprintf("%d of %d members above 90%% memory.", criticalMem, len(members)),
			Container: groupLabel,
			Evidence:  mergeEvidence(evidence, map[string]any{"critical_members": criticalMem}),
			Recommendations: []string{
				"Review limits and heap for the services in this group",
				"Compare with GC and OOM logs from the members",
			},
		})
	} else if pressureCount >= 3 || (pressureCount*100/len(members)) >= 40 {
		out = append(out, model.Insight{
			ID: g.ID + ":group_degradation", Theme: "group_degradation", Severity: "warning",
			Title:   fmt.Sprintf("Simultaneous degradation — %s", groupLabel),
			Summary: fmt.Sprintf("%d members under CPU or memory pressure.", pressureCount),
			Container: groupLabel,
			Evidence:  mergeEvidence(evidence, map[string]any{"pressure_members": pressureCount}),
			Recommendations: []string{
				"Look for a shared dependency (database, queue, gateway)",
				"See the fleet and derived HTTP metrics for the group",
			},
		})
	}

	if errLines >= 20 {
		sev := "warning"
		if errLines >= 100 {
			sev = "critical"
		}
		out = append(out, model.Insight{
			ID: g.ID + ":group_errors", Theme: "error_spike", Severity: sev,
			Title:     fmt.Sprintf("Error spike in group — %s", groupLabel),
			Summary:   fmt.Sprintf("%d error lines across the group in this period.", errLines),
			Container: groupLabel,
			Evidence:  mergeEvidence(evidence, map[string]any{"error_log_lines": errLines}),
			Recommendations: []string{
				"Filter the group logs by the error topic",
				"Follow the trace_id across services",
			},
		})
	}

	if gcLines >= 30 {
		out = append(out, model.Insight{
			ID: g.ID + ":group_gc", Theme: "gc_thrashing", Severity: "warning",
			Title:     fmt.Sprintf("GC pressure in group — %s", groupLabel),
			Summary:   fmt.Sprintf("%d GC lines across the members.", gcLines),
			Container: groupLabel,
			Evidence:  mergeEvidence(evidence, map[string]any{"gc_log_lines": gcLines}),
			Recommendations: []string{
				"Tune the JVM heap for the JVM services in this group",
			},
		})
	}

	restarts, oom := 0, 0
	for _, f := range fleet {
		if _, ok := memberSet[f.Container]; !ok {
			continue
		}
		if f.RestartCount > 3 {
			restarts++
		}
		if f.OOMKilled {
			oom++
		}
	}
	if oom > 0 {
		out = append(out, model.Insight{
			ID: g.ID + ":group_oom", Theme: "oom_killed", Severity: "critical",
			Title:     fmt.Sprintf("OOM kill in group — %s", groupLabel),
			Summary:   fmt.Sprintf("%d members were OOMKilled.", oom),
			Container: groupLabel,
			Evidence:  mergeEvidence(evidence, map[string]any{"oom_members": oom}),
			Recommendations: []string{"Raise the memory limit or reduce the footprint"},
		})
	}
	if restarts >= 2 {
		out = append(out, model.Insight{
			ID: g.ID + ":group_restarts", Theme: "restart_loop", Severity: "warning",
			Title:     fmt.Sprintf("Restart loops in group — %s", groupLabel),
			Summary:   fmt.Sprintf("%d members restarted more than 3 times.", restarts),
			Container: groupLabel,
			Evidence:  mergeEvidence(evidence, map[string]any{"restart_members": restarts}),
			Recommendations: []string{"Verificar fleet e logs de crash dos membros"},
		})
	}

	if hotCPU >= 3 {
		out = append(out, model.Insight{
			ID: g.ID + ":group_cpu", Theme: "cpu_hot", Severity: "warning",
			Title:     fmt.Sprintf("CPU saturated in group — %s", groupLabel),
			Summary:   fmt.Sprintf("%d members at 80%% CPU or more.", hotCPU),
			Container: groupLabel,
			Evidence:  mergeEvidence(evidence, map[string]any{"hot_cpu_members": hotCPU}),
			Recommendations: []string{"Spread the load or add replicas"},
		})
	}

	SortBySeverity(out)
	return out
}

func mergeEvidence(base, extra map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}
