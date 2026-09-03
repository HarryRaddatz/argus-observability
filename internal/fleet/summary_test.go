package fleet

import (
	"testing"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func TestSummarize(t *testing.T) {
	containers := []model.ContainerFleetStatus{
		{Container: "a", State: "running", RestartCount: 2},
		{Container: "b", State: "exited", RestartCount: 1},
		{Container: "c", State: "restarting", Health: "unhealthy", RestartCount: 5},
	}
	s := Summarize(containers)
	if s.Running != 1 || s.Exited != 1 || s.Restarting != 1 || s.Unhealthy != 1 {
		t.Fatalf("unexpected summary: %+v", s)
	}
	if s.TotalRestartCount != 8 {
		t.Fatalf("expected 8 restarts, got %d", s.TotalRestartCount)
	}
}

func TestServiceReplicas(t *testing.T) {
	containers := []model.ContainerFleetStatus{
		{Container: "venuz-api-1", Service: "api", State: "running"},
		{Container: "venuz-api-2", Service: "api", State: "restarting", Health: "unhealthy"},
	}
	services := ServiceReplicas(containers)
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}
	if services[0].ReplicasTotal != 2 || services[0].ReplicasUp != 2 || services[0].Unhealthy != 1 {
		t.Fatalf("unexpected service row: %+v", services[0])
	}
}

func TestBuildResponseUpdatedAt(t *testing.T) {
	ts := time.Now().UTC().Add(-time.Minute)
	resp := BuildResponse([]model.ContainerFleetStatus{
		{Container: "x", State: "running", UpdatedAt: ts},
	}, model.FleetEventStats{})
	if resp.Containers[0].Container != "x" {
		t.Fatal("container missing")
	}
}
