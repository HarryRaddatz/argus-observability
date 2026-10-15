package insights

import (
	"fmt"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

// GenerateFleet builds stability insights from fleet snapshots.
func GenerateFleet(containers []model.ContainerFleetStatus) []model.Insight {
	var out []model.Insight
	seen := map[string]struct{}{}

	for _, c := range containers {
		if c.OOMKilled {
			key := c.Container + ":oom_killed"
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, model.Insight{
				ID: key, Theme: "oom_killed", Severity: "critical",
				Title:     fmt.Sprintf("OOM kill — %s", c.Container),
				Summary:   "The container was killed for running out of memory (OOMKilled).",
				Container: c.Container, EntityUID: c.EntityUID,
				Evidence: map[string]any{"oom_killed": true, "restart_count": c.RestartCount},
				Recommendations: []string{
					"Raise the memory limit in compose or reduce heap and caches",
					"Compare with GC and memory logs before the kill",
				},
			})
		}
		if c.RestartCount > 3 {
			key := c.Container + ":restart_loop"
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			sev := "warning"
			if c.RestartCount > 10 {
				sev = "critical"
			}
			out = append(out, model.Insight{
				ID: key, Theme: "restart_loop", Severity: sev,
				Title:     fmt.Sprintf("Restart loop — %s", c.Container),
				Summary:   fmt.Sprintf("%d restarts so far. The container is unstable or crash looping.", c.RestartCount),
				Container: c.Container, EntityUID: c.EntityUID,
				Evidence: map[string]any{"restart_count": c.RestartCount, "state": c.State},
				Recommendations: []string{
					"Check the startup logs and the exit code",
					"Check the healthcheck and dependencies (database, queues)",
				},
			})
		}
		if c.Health == "unhealthy" {
			key := c.Container + ":unhealthy"
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, model.Insight{
				ID: key, Theme: "unhealthy", Severity: "warning",
				Title:     fmt.Sprintf("Healthcheck failing — %s", c.Container),
				Summary:   "The Docker healthcheck reports unhealthy.",
				Container: c.Container, EntityUID: c.EntityUID,
				Evidence: map[string]any{"health": c.Health, "status_text": c.StatusText},
				Recommendations: []string{
					"Inspect the application health endpoint",
					"Check for latency or an unavailable dependency",
				},
			})
		}
	}
	return out
}
