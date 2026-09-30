package rules

import (
	"testing"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/bus"
	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func TestEvalRuleFiresCPUAfterDuration(t *testing.T) {
	eventBus := bus.New()
	var fired []model.Event
	eventBus.Subscribe(func(evt model.Event) { fired = append(fired, evt) })
	e := NewEngine(eventBus)
	rule := Rule{ID: "cpu-high", Metric: "cpu.usage", Threshold: 80, Duration: time.Minute, Severity: "warning", Title: "CPU high"}
	now := time.Now().UTC()
	labels := model.Labels{"container": "api"}
	e.evalRule(rule, "docker:host:api", "api", labels, 90, now)
	if len(fired) != 0 {
		t.Fatalf("fired too early: %+v", fired)
	}
	e.evalRule(rule, "docker:host:api", "api", labels, 91, now.Add(time.Minute))
	if len(fired) < 1 {
		t.Fatal("expected alert.fired")
	}
	found := false
	for _, evt := range fired {
		if evt.Type == "alert.fired" {
			found = true
		}
	}
	if !found {
		t.Fatalf("events: %+v", fired)
	}
}

func TestEvalRuleResolvesWhenBelowThreshold(t *testing.T) {
	eventBus := bus.New()
	var types []string
	eventBus.Subscribe(func(evt model.Event) { types = append(types, evt.Type) })
	e := NewEngine(eventBus)
	rule := Rule{ID: "memory-high", Metric: "memory.usage_pct", Threshold: 90, Duration: time.Second, Severity: "critical", Title: "Memory"}
	now := time.Now().UTC()
	labels := model.Labels{"container": "api"}
	e.evalRule(rule, "docker:host:api", "api", labels, 95, now)
	e.evalRule(rule, "docker:host:api", "api", labels, 95, now.Add(2*time.Second))
	e.evalRule(rule, "docker:host:api", "api", labels, 10, now.Add(3*time.Second))
	resolved := false
	for _, ty := range types {
		if ty == "alert.resolved" {
			resolved = true
		}
	}
	if !resolved {
		t.Fatalf("types %v", types)
	}
}
