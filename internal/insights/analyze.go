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
				Title:     fmt.Sprintf("Critical memory — %s", w.Container),
				Summary:   fmt.Sprintf("Using %.0f%% of the limit (%.0f MiB of %.0f MiB). An OOM kill is likely soon.", w.MemoryPct, w.MemoryUsage/1024/1024, w.MemoryLimit/1024/1024),
				Container: w.Container, EntityUID: w.EntityUID,
				Evidence: map[string]any{"memory_pct": w.MemoryPct, "cpu_usage": w.CPUUsage},
				Recommendations: []string{
					"Review the memory limit in compose and the application heap",
					"Compare with GC logs and OutOfMemory errors",
					"Consider scaling out or shrinking in-memory caches",
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
				Title:     fmt.Sprintf("High memory — %s", w.Container),
				Summary:   fmt.Sprintf("Sustained use at %.0f%% of the limit.", w.MemoryPct),
				Container: w.Container, EntityUID: w.EntityUID,
				Evidence: map[string]any{"memory_pct": w.MemoryPct},
				Recommendations: []string{
					"Watch the growth trend (RSS against heap)",
					"Filter logs by the GC topic to spot thrashing",
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
				Title:     fmt.Sprintf("CPU saturated — %s", w.Container),
				Summary:   fmt.Sprintf("CPU at %.1f%%. Possible bottleneck or busy loop.", w.CPUUsage),
				Container: w.Container, EntityUID: w.EntityUID,
				Evidence: map[string]any{"cpu_usage": w.CPUUsage},
				Recommendations: []string{
					"Check for blocked threads and profile the CPU",
					"Compare with performance and latency logs",
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
			Title:     fmt.Sprintf("GC pressure — %s", container),
			Summary:   fmt.Sprintf("%d GC lines in this period. Possible thrashing or an undersized heap.", count),
			Container: container, EntityUID: entityByContainer[container],
			Evidence: map[string]any{"gc_log_lines": count},
			Recommendations: []string{
				"Tune the JVM heap (-Xmx/-Xms) or the GC collector (G1/ZGC)",
				"Look for short-lived object allocation",
				"Use the GC log topic to find long pauses",
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
			Title:     fmt.Sprintf("OOM signs — %s", container),
			Summary:   fmt.Sprintf("%d memory or OOM related lines in the logs.", count),
			Container: container, EntityUID: entityByContainer[container],
			Evidence: map[string]any{"oom_log_lines": count},
			Recommendations: []string{
				"First raise the limit or reduce the memory footprint",
				"Look for leaks and caches without a TTL",
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
			Title:     fmt.Sprintf("Error spike — %s", container),
			Summary:   fmt.Sprintf("%d error or exception lines in this period.", count),
			Container: container, EntityUID: entityByContainer[container],
			Evidence: map[string]any{"error_log_lines": count},
			Recommendations: []string{
				"Group repeated stack traces",
				"Check for unavailable external dependencies",
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
