package insights

import "testing"

func TestParseHTTPLogJSONExit(t *testing.T) {
	sig, ok := ParseHTTPLog(`{"ts":"2026-09-03T17:49:57.551+00:00","level":"info","event":"exit","service":"demo-api","status":200,"durationMs":486}`)
	if !ok {
		t.Fatal("expected parse ok")
	}
	if sig.Service != "demo-api" || sig.Status != 200 || sig.DurationMs != 486 {
		t.Fatalf("unexpected sig: %+v", sig)
	}
	if !sig.IsExit {
		t.Fatal("expected exit")
	}
}

func TestParseHTTPLogJSONExitSkipsEntry(t *testing.T) {
	_, ok := ParseHTTPLog(`{"event":"entry","service":"demo-api","msg":"http.request"}`)
	if ok {
		t.Fatal("entry should not produce metrics")
	}
}

func TestParseHTTPLogStructuredResponseLine(t *testing.T) {
	line := `260903/174932.161, (1788457772161:b670754c7016:1:mtlko8vi:11223) [response,api] post /api/orders/filter {"sortBy":"createdAt"} 200 (486ms)`
	sig, ok := ParseHTTPLog(line)
	if !ok {
		t.Fatal("expected parse ok")
	}
	if sig.Method != "POST" || sig.Status != 200 || sig.DurationMs != 486 {
		t.Fatalf("unexpected sig: %+v", sig)
	}
}

func TestParseHTTPLogErrorStatus(t *testing.T) {
	sig, ok := ParseHTTPLog(`{"event":"exit","service":"checkout","status":500,"durationMs":120}`)
	if !ok || sig.Status != 500 {
		t.Fatalf("unexpected: ok=%v sig=%+v", ok, sig)
	}
}

func TestInferServiceFromContainer(t *testing.T) {
	cases := map[string]string{
		"compose-demo-api-1": "demo-api",
		"app-postgres-1":     "postgres",
		"argus-agent":            "argus-agent",
	}
	for in, want := range cases {
		if got := InferServiceFromContainer(in); got != want {
			t.Fatalf("%s: got %q want %q", in, got, want)
		}
	}
}
