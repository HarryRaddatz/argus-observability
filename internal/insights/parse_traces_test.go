package insights

import "testing"

func TestParseTraceFieldsJSON(t *testing.T) {
	msg := `{"level":"info","traceId":"abc123def4567890abcdef1234567890","spanId":"deadbeef","message":"ok"}`
	fields := ParseTraceFields(msg)
	if fields["trace_id"] == "" {
		t.Fatalf("expected trace_id, got %v", fields)
	}
}

func TestParseTraceFieldsTraceparent(t *testing.T) {
	msg := `traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01`
	fields := ParseTraceFields(msg)
	if fields["trace_id"] != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatalf("unexpected trace_id: %v", fields)
	}
	if fields["span_id"] != "00f067aa0ba902b7" {
		t.Fatalf("unexpected span_id: %v", fields)
	}
}

func TestEnrichLogTraceTopic(t *testing.T) {
	_, fields := EnrichLog(`request_id=550e8400-e29b-41d4-a716-446655440000 failed`, "error")
	if fields["trace_id"] == "" {
		t.Fatalf("expected trace_id in fields")
	}
	topics, _ := fields["topics"].([]string)
	found := false
	for _, t := range topics {
		if t == "trace" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected trace topic in %v", topics)
	}
}
