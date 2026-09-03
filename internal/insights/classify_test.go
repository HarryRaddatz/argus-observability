package insights

import "testing"

func TestClassifyLogGC(t *testing.T) {
	topics, signals := ClassifyLog("[GC (Allocation Failure) 45ms]", "info")
	found := false
	for _, tp := range topics {
		if tp == "gc" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected gc topic, got %v", topics)
	}
	if signals["gc_pause_ms"] != float64(45) {
		t.Fatalf("expected pause 45, got %v", signals["gc_pause_ms"])
	}
}

func TestClassifyLogOOM(t *testing.T) {
	topics, _ := ClassifyLog("java.lang.OutOfMemoryError: Java heap space", "error")
	hasOOM := false
	for _, tp := range topics {
		if tp == "oom" || tp == "memory" {
			hasOOM = true
		}
	}
	if !hasOOM {
		t.Fatalf("expected memory/oom topics, got %v", topics)
	}
}
