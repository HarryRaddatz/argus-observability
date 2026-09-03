package docker

import "testing"

func TestCoalesceLoggerAPIStyle(t *testing.T) {
	raw := []parsedLine{
		{message: "[x] Received {"},
		{message: "'$traceId': '6358198d-1280-4b49-9be0-1671bb3ca9a2',"},
		{message: "'$version': 0,"},
		{message: "'$name': 'Responded',"},
		{message: "}"},
	}
	out := coalesceLogLines(raw)
	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d: %+v", len(out), out)
	}
	if !contains(out[0].message, "$traceId") || !contains(out[0].message, "Responded") {
		t.Fatalf("merged message incomplete: %q", out[0].message)
	}
}

func TestCoalesceSplitJSON(t *testing.T) {
	raw := []parsedLine{
		{message: `{"ts":"2026-09-03T17:30:59.931+00:00","level":"info","event":"exit","service":"loggerapi"`},
		{message: "}"},
	}
	out := coalesceLogLines(raw)
	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}
}

func TestCoalesceSeparateEntries(t *testing.T) {
	raw := []parsedLine{
		{message: `{"level":"info","event":"entry"}`},
		{message: `{"level":"info","event":"exit"}`},
	}
	out := coalesceLogLines(raw)
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
