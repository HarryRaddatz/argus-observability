package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func TestTopologyLinksPreferObservedOverInferred(t *testing.T) {
	st := openTest(t)

	now := time.Now().UTC().Truncate(time.Second)
	logs := []model.LogEntry{{
		TS:      now,
		Message: `calling http://gateway/checkout`,
		Labels:  model.Labels{"service": "checkout"},
	}}
	if err := st.RecordTopologyEdges(context.Background(), logs); err != nil {
		t.Fatal(err)
	}

	links := []model.TopologyLink{
		{Source: "checkout", Target: "payments", Kind: "http", Port: 8080, Count: 12, TS: now},
		{Source: "checkout", Target: "gateway", Kind: "tcp", Port: 9000, Count: 4, TS: now},
		{Source: "", Target: "nowhere", Count: 1, TS: now},
	}
	if err := st.RecordTopologyLinks(context.Background(), links); err != nil {
		t.Fatal(err)
	}
	// A second window must accumulate the count, not replace it.
	links[0].Count = 8
	if err := st.RecordTopologyLinks(context.Background(), links[:1]); err != nil {
		t.Fatal(err)
	}

	graph, err := st.GetTopology(context.Background(), now.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	byTarget := map[string]model.TopologyEdge{}
	for _, edge := range graph.Edges {
		byTarget[edge.Target] = edge
	}
	if got := byTarget["payments"]; got.Count != 20 || got.Origin != model.TopologyOriginKernel || got.Port != 8080 {
		t.Fatalf("observed edge: %+v", got)
	}
	if got := byTarget["gateway"]; got.Origin != model.TopologyOriginKernel {
		t.Fatalf("kernel edge must win over the one inferred from the log line: %+v", got)
	}
	if _, ok := byTarget["nowhere"]; ok {
		t.Fatal("links without a source must be dropped")
	}
}
