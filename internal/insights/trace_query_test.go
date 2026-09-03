package insights

import "testing"

func TestTraceSearchPatternsUUID(t *testing.T) {
	patterns := TraceSearchPatterns("50e30959-f59e-487f-bbe0-89ae7d8e74e5")
	if len(patterns) < 3 {
		t.Fatalf("expected multiple patterns, got %v", patterns)
	}
	found := false
	for _, p := range patterns {
		if p == `%50e30959-f59e-487f-bbe0-89ae7d8e74e5%` {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing dashed uuid pattern: %v", patterns)
	}
}

func TestTraceSearchPatternsHex(t *testing.T) {
	patterns := TraceSearchPatterns("50e30959f59e487fbbe089ae7d8e74e5")
	if len(patterns) < 2 {
		t.Fatalf("expected patterns for hex trace, got %v", patterns)
	}
}
