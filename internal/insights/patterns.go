package insights

import (
	"fmt"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

// GeneratePatternSpikes emits insights for frequently repeated log patterns.
func GeneratePatternSpikes(patterns []model.LogPattern) []model.Insight {
	var out []model.Insight
	for _, p := range patterns {
		if p.Count < 50 {
			continue
		}
		sev := "warning"
		if p.Count >= 200 {
			sev = "critical"
		}
		id := "pattern:" + p.PatternKey + ":" + p.Container
		label := p.Service
		if label == "" {
			label = p.Container
		}
		out = append(out, model.Insight{
			ID: id, Theme: "log_pattern_spike", Severity: sev,
			Title:     fmt.Sprintf("Padrão repetido — %s", label),
			Summary:   fmt.Sprintf("%d ocorrências: %s", p.Count, truncate(p.Pattern, 120)),
			Container: p.Container,
			Evidence: map[string]any{
				"pattern_key": p.PatternKey, "count": p.Count, "sample": p.Sample,
			},
			Recommendations: []string{
				"Investigar causa raiz do padrão repetido",
				"Usar drill-down em /logs/patterns",
			},
		})
	}
	return out
}

// GenerateTopologyChain detects downstream error spikes on topology edges.
func GenerateTopologyChain(graph model.TopologyGraph, logStats []LogTopicStats) []model.Insight {
	errByService := map[string]int{}
	for _, s := range logStats {
		if s.Topic != "error" {
			continue
		}
		svc := s.Container
		errByService[svc] += s.Count
	}
	var out []model.Insight
	for _, e := range graph.Edges {
		if errByService[e.Target] < 10 {
			continue
		}
		out = append(out, model.Insight{
			ID: fmt.Sprintf("topo:%s->%s", e.Source, e.Target), Theme: "chain_degradation", Severity: "warning",
			Title:     fmt.Sprintf("Degradação em cadeia — %s → %s", e.Source, e.Target),
			Summary:   fmt.Sprintf("Destino com %d erros; %d chamadas observadas (%s).", errByService[e.Target], e.Count, e.Kind),
			Container: e.Target,
			Evidence: map[string]any{
				"source": e.Source, "target": e.Target, "kind": e.Kind,
				"edge_count": e.Count, "target_errors": errByService[e.Target],
			},
			Recommendations: []string{
				"Verificar latência upstream e timeouts downstream",
				"Abrir mapa de topologia e logs do serviço destino",
			},
		})
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
