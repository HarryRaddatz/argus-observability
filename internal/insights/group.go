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
			Title:   fmt.Sprintf("Memória crítica no grupo — %s", groupLabel),
			Summary: fmt.Sprintf("%d de %d membros acima de 90%% de memória.", criticalMem, len(members)),
			Container: groupLabel,
			Evidence:  mergeEvidence(evidence, map[string]any{"critical_members": criticalMem}),
			Recommendations: []string{
				"Revisar limites e heap dos serviços do grupo",
				"Correlacionar com logs GC/OOM dos membros",
			},
		})
	} else if pressureCount >= 3 || (pressureCount*100/len(members)) >= 40 {
		out = append(out, model.Insight{
			ID: g.ID + ":group_degradation", Theme: "group_degradation", Severity: "warning",
			Title:   fmt.Sprintf("Degradação simultânea — %s", groupLabel),
			Summary: fmt.Sprintf("%d membros com pressão de CPU/memória no grupo.", pressureCount),
			Container: groupLabel,
			Evidence:  mergeEvidence(evidence, map[string]any{"pressure_members": pressureCount}),
			Recommendations: []string{
				"Investigar dependência compartilhada (DB, fila, gateway)",
				"Ver fleet e métricas HTTP derivadas do grupo",
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
			Title:     fmt.Sprintf("Pico de erros no grupo — %s", groupLabel),
			Summary:   fmt.Sprintf("%d linhas de erro agregadas no período.", errLines),
			Container: groupLabel,
			Evidence:  mergeEvidence(evidence, map[string]any{"error_log_lines": errLines}),
			Recommendations: []string{
				"Filtrar logs do grupo por tópico error",
				"Correlacionar com traceId entre serviços",
			},
		})
	}

	if gcLines >= 30 {
		out = append(out, model.Insight{
			ID: g.ID + ":group_gc", Theme: "gc_thrashing", Severity: "warning",
			Title:     fmt.Sprintf("Pressão de GC no grupo — %s", groupLabel),
			Summary:   fmt.Sprintf("%d linhas GC agregadas entre os membros.", gcLines),
			Container: groupLabel,
			Evidence:  mergeEvidence(evidence, map[string]any{"gc_log_lines": gcLines}),
			Recommendations: []string{
				"Ajustar heap JVM nos serviços JVM do grupo",
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
			Title:     fmt.Sprintf("OOM kill no grupo — %s", groupLabel),
			Summary:   fmt.Sprintf("%d membros com OOMKilled.", oom),
			Container: groupLabel,
			Evidence:  mergeEvidence(evidence, map[string]any{"oom_members": oom}),
			Recommendations: []string{"Aumentar limite de memória ou reduzir footprint"},
		})
	}
	if restarts >= 2 {
		out = append(out, model.Insight{
			ID: g.ID + ":group_restarts", Theme: "restart_loop", Severity: "warning",
			Title:     fmt.Sprintf("Restart loops no grupo — %s", groupLabel),
			Summary:   fmt.Sprintf("%d membros com RestartCount > 3.", restarts),
			Container: groupLabel,
			Evidence:  mergeEvidence(evidence, map[string]any{"restart_members": restarts}),
			Recommendations: []string{"Verificar fleet e logs de crash dos membros"},
		})
	}

	if hotCPU >= 3 {
		out = append(out, model.Insight{
			ID: g.ID + ":group_cpu", Theme: "cpu_hot", Severity: "warning",
			Title:     fmt.Sprintf("CPU saturada no grupo — %s", groupLabel),
			Summary:   fmt.Sprintf("%d membros com CPU >= 80%%.", hotCPU),
			Container: groupLabel,
			Evidence:  mergeEvidence(evidence, map[string]any{"hot_cpu_members": hotCPU}),
			Recommendations: []string{"Distribuir carga ou escalar réplicas"},
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
