package slo

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/bus"
	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/HarryRaddatz/argus-observability/internal/store"
	"github.com/google/uuid"
)

type Evaluator struct {
	bus    *bus.Bus
	dedupe time.Duration
	last   map[string]time.Time
}

func NewEvaluator(eventBus *bus.Bus) *Evaluator {
	return &Evaluator{bus: eventBus, dedupe: 30 * time.Minute, last: map[string]time.Time{}}
}

func (e *Evaluator) Evaluate(ctx context.Context, st store.Store) {
	slos, err := st.ListSLOs(ctx)
	if err != nil {
		return
	}
	now := time.Now().UTC()
	for _, def := range slos {
		status, err := st.EvaluateSLO(ctx, def, now)
		if err != nil {
			continue
		}
		if status.ErrorBudgetRemaining >= 10 {
			continue
		}
		key := def.ID + ":budget"
		if t, ok := e.last[key]; ok && now.Sub(t) < e.dedupe {
			continue
		}
		e.last[key] = now
		entity := def.Service
		if entity == "" {
			entity = def.GroupID
		}
		e.bus.Publish(model.Event{
			ID: uuid.NewString(), Type: "slo.budget_low", TS: now, Severity: "warning",
			Source: "slo-engine", EntityUID: "slo:" + def.ID,
			Labels: model.Labels{"service": def.Service, "slo_id": def.ID},
			Payload: map[string]any{
				"slo_id": def.ID, "name": def.Name, "compliance": status.Compliance,
				"error_budget_remaining": status.ErrorBudgetRemaining,
				"p95_latency_ms": status.P95LatencyMs,
			},
		})
	}
}

func P95(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	cp := append([]float64(nil), values...)
	sort.Float64s(cp)
	idx := int(math.Ceil(float64(len(cp))*0.95)) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(cp) {
		idx = len(cp) - 1
	}
	return cp[idx]
}

func ComplianceLatency(latencies []float64, thresholdMs float64, targetPct float64) float64 {
	if len(latencies) == 0 {
		return 100
	}
	good := 0
	for _, d := range latencies {
		if d <= thresholdMs {
			good++
		}
	}
	return (float64(good) / float64(len(latencies))) * 100
}

func ErrorBudgetRemaining(compliance, targetPct float64) float64 {
	if targetPct <= 0 {
		return 100
	}
	if compliance >= targetPct {
		return 100
	}
	allowedBad := 100 - targetPct
	actualBad := 100 - compliance
	if allowedBad <= 0 {
		return 0
	}
	remaining := (allowedBad - actualBad) / allowedBad * 100
	if remaining < 0 {
		return 0
	}
	if remaining > 100 {
		return 100
	}
	return remaining
}
