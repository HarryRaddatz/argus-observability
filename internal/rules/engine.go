package rules

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/bus"
	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/HarryRaddatz/argus-observability/internal/store"
	"github.com/google/uuid"
)

type Rule struct {
	ID        string
	Metric    string
	Threshold float64
	Duration  time.Duration
	Severity  string
	Title     string
}

type Engine struct {
	bus     *bus.Bus
	rules   []Rule
	dedupe  time.Duration
	mu      sync.Mutex
	track   map[string]*tracker
	active  map[string]ActiveAlert
}

type ActiveAlert struct {
	RuleID    string
	EntityUID string
	Container string
	Title     string
	Severity  string
	Summary   string
	FiredAt   time.Time
	Value     float64
}

type tracker struct {
	since time.Time
	fired bool
}

func NewEngine(eventBus *bus.Bus) *Engine {
	return &Engine{
		bus:    eventBus,
		dedupe: 5 * time.Minute,
		track:  map[string]*tracker{},
		active: map[string]ActiveAlert{},
		rules: []Rule{
			{ID: "cpu-high", Metric: "cpu.usage", Threshold: 80, Duration: 5 * time.Minute, Severity: "warning", Title: "CPU elevada"},
			{ID: "memory-high", Metric: "memory.usage_pct", Threshold: 90, Duration: 5 * time.Minute, Severity: "critical", Title: "Memória crítica"},
		},
	}
}

func (e *Engine) Evaluate(ctx context.Context, st store.Store) {
	workloads, err := st.ListWorkloads(ctx, time.Now().UTC().Add(-10*time.Minute))
	if err != nil {
		return
	}
	now := time.Now().UTC()
	for _, w := range workloads {
		pct := 0.0
		if w.MemoryLimit > 0 {
			pct = (w.MemoryUsage / w.MemoryLimit) * 100
		}
		for _, rule := range e.rules {
			val := w.CPUUsage
			if rule.Metric == "memory.usage_pct" {
				val = pct
			}
			e.evalRule(rule, w.EntityUID, w.Container, w.Labels, val, now)
		}
	}
}

func (e *Engine) evalRule(rule Rule, entityUID, container string, labels model.Labels, value float64, now time.Time) {
	key := rule.ID + ":" + entityUID
	above := value >= rule.Threshold

	e.mu.Lock()
	defer e.mu.Unlock()

	tr := e.track[key]
	if tr == nil {
		tr = &tracker{}
		e.track[key] = tr
	}

	if above {
		if tr.since.IsZero() {
			tr.since = now
		}
		if !tr.fired && now.Sub(tr.since) >= rule.Duration {
			tr.fired = true
			e.fire(rule, entityUID, container, labels, value, now)
		}
		return
	}

	if tr.fired {
		e.resolve(rule, entityUID, container, labels, value, now)
	}
	delete(e.track, key)
}

func (e *Engine) fire(rule Rule, entityUID, container string, labels model.Labels, value float64, now time.Time) {
	key := rule.ID + ":" + entityUID
	e.active[key] = ActiveAlert{
		RuleID: rule.ID, EntityUID: entityUID, Container: container,
		Title: fmt.Sprintf("%s — %s", rule.Title, container), Severity: rule.Severity,
		Summary: fmt.Sprintf("%s em %.1f (limiar %.0f)", rule.Metric, value, rule.Threshold),
		FiredAt: now, Value: value,
	}
	e.bus.Publish(model.Event{
		ID: uuid.NewString(), Type: "alert.fired", TS: now, Severity: rule.Severity,
		Source: "rule-engine", EntityUID: entityUID, Labels: mergeLabels(labels, rule.ID),
		Payload: map[string]any{
			"rule_id": rule.ID, "metric": rule.Metric, "value": value, "threshold": rule.Threshold,
		},
	})
	e.bus.Publish(model.Event{
		ID: uuid.NewString(), Type: "metric.threshold", TS: now, Severity: rule.Severity,
		Source: "rule-engine", EntityUID: entityUID, Labels: mergeLabels(labels, rule.ID),
		Payload: map[string]any{
			"rule_id": rule.ID, "metric": rule.Metric, "value": value, "threshold": rule.Threshold,
		},
	})
}

func (e *Engine) resolve(rule Rule, entityUID, container string, labels model.Labels, value float64, now time.Time) {
	key := rule.ID + ":" + entityUID
	delete(e.active, key)
	e.bus.Publish(model.Event{
		ID: uuid.NewString(), Type: "alert.resolved", TS: now, Severity: "info",
		Source: "rule-engine", EntityUID: entityUID, Labels: mergeLabels(labels, rule.ID),
		Payload: map[string]any{
			"rule_id": rule.ID, "metric": rule.Metric, "value": value, "threshold": rule.Threshold,
		},
	})
}

func (e *Engine) ActiveInsights() []model.Insight {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]model.Insight, 0, len(e.active))
	for key, a := range e.active {
		out = append(out, model.Insight{
			ID: "alert:" + key, Theme: "alert_active", Severity: a.Severity,
			Title: a.Title, Summary: a.Summary, Container: a.Container, EntityUID: a.EntityUID,
			Evidence: map[string]any{
				"rule_id": a.RuleID, "fired_at": a.FiredAt.Format(time.RFC3339), "value": a.Value,
			},
			Recommendations: []string{"Alerta ativo — ver métricas e logs do container"},
		})
	}
	return out
}

func (e *Engine) ActiveAlerts() []ActiveAlert {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]ActiveAlert, 0, len(e.active))
	for _, a := range e.active {
		out = append(out, a)
	}
	return out
}

func mergeLabels(labels model.Labels, ruleID string) model.Labels {
	out := model.Labels{"rule": ruleID}
	for k, v := range labels {
		out[k] = v
	}
	return out
}
