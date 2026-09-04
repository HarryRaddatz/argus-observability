package insights

import (
	"testing"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func TestGenerateFleetOOM(t *testing.T) {
	out := GenerateFleet([]model.ContainerFleetStatus{
		{Container: "stack-demo-api", EntityUID: "docker:host:stack-demo-api", OOMKilled: true},
	})
	if len(out) != 1 || out[0].Theme != "oom_killed" {
		t.Fatalf("expected oom insight, got %+v", out)
	}
}

func TestGenerateFleetRestartLoop(t *testing.T) {
	out := GenerateFleet([]model.ContainerFleetStatus{
		{Container: "stack-demo-api", EntityUID: "x", RestartCount: 5},
	})
	if len(out) != 1 || out[0].Theme != "restart_loop" {
		t.Fatalf("expected restart insight, got %+v", out)
	}
}
