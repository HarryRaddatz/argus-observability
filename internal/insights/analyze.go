package insights

import (
	"fmt"
	"strings"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

// LogTopicStats aggregates classified log volume per container.
type LogTopicStats struct {
	Container string
	EntityUID string
	Topic     string
	Count     int
}

// WorkloadMemory holds memory pressure context for insight generation.
type WorkloadMemory struct {
	Container   string
	EntityUID   string
	CPUUsage    float64
	MemoryPct   float64
	MemoryUsage float64
	MemoryLimit float64
}

// Generate builds optimization insights from workloads and log topic stats.
func Generate(workloads []WorkloadMemory, logStats []LogTopicStats) []model.Insight {
	var out []model.Insight
	seen := map[string]struct{}{}

	for _, w := range workloads {
		if w.MemoryPct >= 90 {
			key := w.Container + ":memory_critical"
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, model.Insight{
				ID: key, Theme: "memory_pressure", Severity: "critical",
				Title:     fmt.Sprintf("Memória crítica — %s", w.Container),
				Summary:   fmt.Sprintf("Uso em %.0f%% do limite (%.0f MiB / %.0f MiB). Risco iminente de OOM.", w.MemoryPct, w.MemoryUsage/1024/1024, w.MemoryLimit/1024/1024),
				Container: w.Container, EntityUID: w.EntityUID,
				Evidence: map[string]any{"memory_pct": w.MemoryPct, "cpu_usage": w.CPUUsage},
				Recommendations: []string{
					"Revisar limites de memória no compose e heap da aplicação",
					"Correlacionar com logs de GC e erros OutOfMemory",
					"Considerar scale horizontal ou redução de cache in-memory",
				},
			})
		} else if w.MemoryPct >= 75 {
			key := w.Container + ":memory_high"
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, model.Insight{
				ID: key, Theme: "memory_pressure", Severity: "warning",
				Title:     fmt.Sprintf("Memória elevada — %s", w.Container),
				Summary:   fmt.Sprintf("Uso sustentado em %.0f%% do limite.", w.MemoryPct),
				Container: w.Container, EntityUID: w.EntityUID,
				Evidence: map[string]any{"memory_pct": w.MemoryPct},
				Recommendations: []string{
					"Monitorar tendência de crescimento (RSS vs heap)",
					"Filtrar logs por tópico GC para detectar thrashing",
				},
			})
		}
		if w.CPUUsage >= 80 {
			key := w.Container + ":cpu_hot"
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, model.Insight{
				ID: key, Theme: "cpu_hot", Severity: "warning",
				Title:     fmt.Sprintf("CPU saturada — %s", w.Container),
				Summary:   fmt.Sprintf("CPU em %.1f%% — possível gargalo ou loop ativo.", w.CPUUsage),
				Container: w.Container, EntityUID: w.EntityUID,
				Evidence: map[string]any{"cpu_usage": w.CPUUsage},
				Recommendations: []string{
					"Verificar threads bloqueadas e profiling de CPU",
					"Correlacionar com logs de performance/latência",
				},
			})
		}
	}

	gcByContainer := map[string]int{}
	oomByContainer := map[string]int{}
	errByContainer := map[string]int{}
	entityByContainer := map[string]string{}
	for _, s := range logStats {
		entityByContainer[s.Container] = s.EntityUID
		switch s.Topic {
		case "gc":
			gcByContainer[s.Container] += s.Count
		case "oom":
			oomByContainer[s.Container] += s.Count
		case "memory":
			oomByContainer[s.Container] += s.Count
		case "error":
			errByContainer[s.Container] += s.Count
		}
	}

	for container, count := range gcByContainer {
		if count < 15 {
			continue
		}
		key := container + ":gc_thrash"
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		sev := "warning"
		if count >= 50 {
			sev = "critical"
		}
		out = append(out, model.Insight{
			ID: key, Theme: "gc_thrashing", Severity: sev,
			Title:     fmt.Sprintf("Pressão de GC — %s", container),
			Summary:   fmt.Sprintf("%d linhas de GC no período — possível thrashing ou heap subdimensionado.", count),
			Container: container, EntityUID: entityByContainer[container],
			Evidence: map[string]any{"gc_log_lines": count},
			Recommendations: []string{
				"Ajustar heap JVM (-Xmx/-Xms) ou GC collector (G1/ZGC)",
				"Investigar alocação de objetos de curta duração",
				"Usar filtro de logs tópico GC para ver pausas longas",
			},
		})
	}

	for container, count := range oomByContainer {
		if count == 0 {
			continue
		}
		key := container + ":oom"
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, model.Insight{
			ID: key, Theme: "oom_risk", Severity: "critical",
			Title:     fmt.Sprintf("Sinais de OOM — %s", container),
			Summary:   fmt.Sprintf("%d ocorrências relacionadas a memória/OOM nos logs.", count),
			Container: container, EntityUID: entityByContainer[container],
			Evidence: map[string]any{"oom_log_lines": count},
			Recommendations: []string{
				"Prioridade: aumentar limite ou reduzir footprint",
				"Revisar vazamentos e caches sem TTL",
			},
		})
	}

	for container, count := range errByContainer {
		if count < 10 {
			continue
		}
		key := container + ":errors"
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, model.Insight{
			ID: key, Theme: "error_spike", Severity: "warning",
			Title:     fmt.Sprintf("Pico de erros — %s", container),
			Summary:   fmt.Sprintf("%d linhas de erro/exception no período.", count),
			Container: container, EntityUID: entityByContainer[container],
			Evidence: map[string]any{"error_log_lines": count},
			Recommendations: []string{
				"Agrupar stack traces repetidos",
				"Verificar dependências externas indisponíveis",
			},
		})
	}

	// Sort by severity weight
	SortBySeverity(out)
	return out
}

// SortBySeverity orders insights critical first.
func SortBySeverity(out []model.Insight) {
	order := map[string]int{"critical": 0, "warning": 1, "info": 2}
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if order[out[j].Severity] < order[out[i].Severity] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
}

func ContainerFromEntity(entityUID string) string {
	parts := strings.Split(entityUID, ":")
	if len(parts) == 0 {
		return entityUID
	}
	return parts[len(parts)-1]
}
