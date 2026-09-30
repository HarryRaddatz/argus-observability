package slo

import "testing"

func TestP95(t *testing.T) {
	if P95(nil) != 0 {
		t.Fatal("empty")
	}
	got := P95([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	if got < 9.5 || got > 10 {
		t.Fatalf("p95=%v", got)
	}
}

func TestErrorBudgetRemaining(t *testing.T) {
	if ErrorBudgetRemaining(99.9, 99.9) != 100 {
		t.Fatal("at target")
	}
	if ErrorBudgetRemaining(99, 99.9) != 0 {
		t.Fatal("below target is exhausted")
	}
	if ErrorBudgetRemaining(0, 99.9) != 0 {
		t.Fatal("exhausted")
	}
}

func TestComplianceLatency(t *testing.T) {
	if ComplianceLatency(nil, 200, 99) != 100 {
		t.Fatal("empty is 100")
	}
	got := ComplianceLatency([]float64{10, 20, 300}, 200, 99)
	if got < 66 || got > 67 {
		t.Fatalf("compliance=%v", got)
	}
}
