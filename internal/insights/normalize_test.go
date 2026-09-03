package insights

import "testing"

func TestNormalizeLogMessage(t *testing.T) {
	raw := `2026-09-03T17:49:08.416+00:00 warn No handlers traceId=50e30959-f59e-487f-bbe0-89ae7d8e74e5 count=42`
	norm := NormalizeLogMessage(raw)
	if norm == raw {
		t.Fatalf("expected normalization, got %q", norm)
	}
	if !containsAll(norm, "<ts>", "<uuid>", "<n>") {
		t.Fatalf("missing placeholders: %q", norm)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !contains(s, p) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
