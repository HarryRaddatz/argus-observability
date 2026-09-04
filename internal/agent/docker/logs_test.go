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

func TestCoalesceDoesNotMergeHTTPWithStructuredJSON(t *testing.T) {
	httpLine := `260903/174902.099, (1788457742099:b670754c7016:1:mtlko8vi:11216) [response,api] http://b670754c7016:9002: post /api/orders/foo {} 200 (2026ms)`
	jsonLine := `{"ts":"2026-09-03T17:49:08.416+00:00","level":"info","event":"entry","service":"demo-api","traceId":"50e30959-f59e-487f-bbe0-89ae7d8e74e5","msg":"http.request"}`
	raw := []parsedLine{
		{message: httpLine},
		{message: jsonLine},
		{message: `{"ts":"2026-09-03T17:49:08.420+00:00","level":"info","event":"exit","service":"demo-api"}`},
	}
	out := coalesceLogLines(raw)
	if len(out) != 3 {
		t.Fatalf("expected 3 entries, got %d: messages=%v", len(out), out)
	}
}

func TestCoalesceDoesNotMergeMultipleStructuredJSON(t *testing.T) {
	raw := []parsedLine{
		{message: `{"ts":"2026-09-03T17:49:02.099+00:00","level":"info","event":"entry","service":"demo-api"}`},
		{message: `{"ts":"2026-09-03T17:49:02.106+00:00","level":"info","event":"info","service":"demo-api"}`},
		{message: `{"ts":"2026-09-03T17:49:02.516+00:00","level":"info","event":"exit","service":"demo-api"}`},
	}
	out := coalesceLogLines(raw)
	if len(out) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(out))
	}
}

func TestCoalesceWinstonMultilineThenSeparate(t *testing.T) {
	raw := []parsedLine{
		{message: `2026-09-03T17:50:56.147Z warn: No handlers were registered for message. {`},
		{message: `"name": "HandlerRegistry",`},
		{message: `"receivedMessage": {`},
		{message: `"$traceId": "a443c602-8746-43b0-b0e7-121c28b4ea88"`},
		{message: `}`},
		{message: `}`},
		{message: `2026-09-03T17:50:56.147Z warn: No handlers registered for message. Message will be discarded {`},
		{message: `"name": "ServiceBus",`},
		{message: `"messageType": {`},
		{message: `"$body": false,`},
		{message: `"$traceId": "a443c602-8746-43b0-b0e7-121c28b4ea88"`},
		{message: `}`},
		{message: `}`},
	}
	out := coalesceLogLines(raw)
	if len(out) != 2 {
		t.Fatalf("expected 2 winston entries, got %d", len(out))
	}
}

func TestCoalesceStructuredBurst(t *testing.T) {
	var raw []parsedLine
	msgs := []string{
		`{"ts":"2026-09-03T17:49:32.647+00:00","level":"info","event":"exit","service":"demo-api","traceId":"bbf65b54","status":200}`,
		`260903/174932.161, (1788457772161:b670754c7016:1:mtlko8vi:11223) [response,api] post /api/orders/filter {"sortBy":"createdAt","orderBy":"desc"} 200 (486ms)`,
		`{"ts":"2026-09-03T17:49:56.915+00:00","level":"info","event":"entry","service":"demo-api"}`,
		`{"ts":"2026-09-03T17:49:56.928+00:00","level":"info","event":"entry","service":"demo-api"}`,
		`{"ts":"2026-09-03T17:49:57.551+00:00","level":"info","event":"exit","service":"demo-api","status":200}`,
	}
	for _, m := range msgs {
		raw = append(raw, parsedLine{message: m})
	}
	out := coalesceLogLines(raw)
	if len(out) != len(msgs) {
		t.Fatalf("expected %d entries, got %d", len(msgs), len(out))
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
