package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func TestListTracesMergesSpansAndLogs(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)

	spans := []model.TraceSpan{
		{TraceID: "aaaa1111", SpanID: "s1", Name: "GET /orders", Service: "orders", Container: "shop-orders-1",
			StartTS: now.Add(-2 * time.Second), EndTS: now.Add(-1 * time.Second), Status: "ok", Kind: "server", Source: "otlp"},
		{TraceID: "aaaa1111", SpanID: "s2", ParentSpanID: "s1", Name: "SELECT orders", Service: "orders-db", Container: "shop-db-1",
			StartTS: now.Add(-1900 * time.Millisecond), EndTS: now.Add(-1500 * time.Millisecond), Status: "error", Kind: "client", Source: "otlp"},
	}
	if err := st.WriteTraceSpans(ctx, spans); err != nil {
		t.Fatal(err)
	}
	logs := []model.LogEntry{
		{TS: now.Add(-30 * time.Second), Message: "payment started", Level: "info", EntityUID: "docker:host:shop-payments-1",
			Labels: model.Labels{"container": "shop-payments-1"}, Fields: map[string]any{"trace_id": "bbbb2222"}},
		{TS: now.Add(-29 * time.Second), Message: "payment failed", Level: "error", EntityUID: "docker:host:shop-payments-1",
			Labels: model.Labels{"container": "shop-payments-1"}, Fields: map[string]any{"trace_id": "bbbb2222"}},
		{TS: now.Add(-3 * time.Second), Message: "orders log line", Level: "info", EntityUID: "docker:host:shop-orders-1",
			Labels: model.Labels{"container": "shop-orders-1"}, Fields: map[string]any{"traceId": "AAAA1111"}},
		{TS: now.Add(-3 * time.Hour), Message: "too old", Level: "info", EntityUID: "docker:host:shop-orders-1",
			Labels: model.Labels{"container": "shop-orders-1"}, Fields: map[string]any{"trace_id": "cccc3333"}},
	}
	if err := st.WriteLogs(ctx, logs); err != nil {
		t.Fatal(err)
	}

	got, err := st.ListTraces(ctx, model.TraceListFilter{Since: now.Add(-time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 traces, got %d: %+v", len(got), got)
	}
	otlp, fromLogs := got[0], got[1]
	if otlp.TraceID != "aaaa1111" || otlp.Source != "otlp" || otlp.SpanCount != 2 || !otlp.Error {
		t.Fatalf("unexpected otlp trace: %+v", otlp)
	}
	if otlp.Name != "GET /orders" || otlp.Service != "orders" {
		t.Fatalf("otlp trace should be named after its root span: %+v", otlp)
	}
	if otlp.DurationMs != 1000 {
		t.Fatalf("otlp duration = %v, want 1000", otlp.DurationMs)
	}
	if fromLogs.TraceID != "bbbb2222" || fromLogs.Source != "logs" || fromLogs.SpanCount != 2 || !fromLogs.Error {
		t.Fatalf("unexpected log trace: %+v", fromLogs)
	}
	if fromLogs.Name != "payment started" || fromLogs.Service != "payments" || fromLogs.Container != "shop-payments-1" {
		t.Fatalf("log trace should use its earliest line and inferred service: %+v", fromLogs)
	}

	filtered, err := st.ListTraces(ctx, model.TraceListFilter{Since: now.Add(-time.Hour), Service: "orders-db"})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || filtered[0].TraceID != "aaaa1111" {
		t.Fatalf("service filter should match any span of the trace: %+v", filtered)
	}

	for msg, want := range map[string]string{
		`{"event":"exit","method":"GET","path":"/orders","status":200}`: "GET /orders",
		`{"msg":"query orders","trace_id":"x"}`:                         "query orders",
		"plain line":                                                    "plain line",
	} {
		if got := logTraceName(msg); got != want {
			t.Fatalf("logTraceName(%q) = %q, want %q", msg, got, want)
		}
	}

	limited, err := st.ListTraces(ctx, model.TraceListFilter{Since: now.Add(-time.Hour), Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(limited) != 1 || limited[0].TraceID != "aaaa1111" {
		t.Fatalf("limit should keep the most recent trace: %+v", limited)
	}
}
