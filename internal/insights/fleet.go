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
				Summary:   "Container foi encerrado por falta de memória (OOMKilled).",
				Container: c.Container, EntityUID: c.EntityUID,
				Evidence: map[string]any{"oom_killed": true, "restart_count": c.RestartCount},
				Recommendations: []string{
					"Aumentar limite de memória no compose ou reduzir heap/cache",
					"Correlacionar com logs de GC e memória antes do kill",
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
				Summary:   fmt.Sprintf("%d restarts acumulados — instabilidade ou crash loop.", c.RestartCount),
				Container: c.Container, EntityUID: c.EntityUID,
				Evidence: map[string]any{"restart_count": c.RestartCount, "state": c.State},
				Recommendations: []string{
					"Verificar logs de startup e exit code",
					"Conferir healthcheck e dependências (DB, filas)",
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
				Title:     fmt.Sprintf("Healthcheck falhou — %s", c.Container),
				Summary:   "Docker healthcheck reportou unhealthy.",
				Container: c.Container, EntityUID: c.EntityUID,
				Evidence: map[string]any{"health": c.Health, "status_text": c.StatusText},
				Recommendations: []string{
					"Inspecionar endpoint de health da aplicação",
					"Verificar latência ou dependência indisponível",
				},
			})
		}
	}
	return out
}
