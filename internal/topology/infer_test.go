package topology

import (
	"testing"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func TestInferTopologyEdgesHTTP(t *testing.T) {
	edges := InferTopologyEdges(model.LogEntry{
		Message: `GET http://payments:8080/charge`,
		Labels:  model.Labels{"service": "checkout", "container": "checkout"},
	})
	if len(edges) == 0 {
		t.Fatal("expected edges")
	}
	found := false
	for _, e := range edges {
		if e.Source == "checkout" && e.Kind == "http" && e.Target != "" {
			found = true
		}
	}
	if !found {
		t.Fatalf("%+v", edges)
	}
}

func TestInferTopologyEdgesNoService(t *testing.T) {
	edges := InferTopologyEdges(model.LogEntry{Message: "http://other"})
	if len(edges) != 0 {
		t.Fatalf("got %v", edges)
	}
}
