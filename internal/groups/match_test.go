package groups

import (
	"testing"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func TestMatchesAndDiscover(t *testing.T) {
	w := model.WorkloadSnapshot{
		Container: "api",
		Service:   "checkout",
		Stack:     "shop",
		Labels:    model.Labels{"service": "checkout", "stack": "shop"},
	}
	if !Matches(w, model.WorkloadGroup{Kind: model.GroupKindService, LabelValue: "checkout"}) {
		t.Fatal("service match")
	}
	if !Matches(w, model.WorkloadGroup{Kind: model.GroupKindStack, LabelValue: "shop"}) {
		t.Fatal("stack match")
	}
	if !Matches(w, model.WorkloadGroup{Kind: model.GroupKindCustom, Containers: []string{"api"}}) {
		t.Fatal("custom match")
	}
	got := Discover([]model.WorkloadSnapshot{w})
	if len(got) == 0 {
		t.Fatal("expected discovered groups")
	}
}
