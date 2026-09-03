package insights

import (
	"regexp"
	"strings"
)

var (
	traceParentRe = regexp.MustCompile(`(?i)traceparent[=:\s]+00-([0-9a-f]{32})-([0-9a-f]{16})`)
	otelTraceJSONRe = regexp.MustCompile(`(?i)"trace[_-]?id"\s*:\s*"([0-9a-fA-F-]{16,64})"`)
	otelSpanJSONRe  = regexp.MustCompile(`(?i)"span[_-]?id"\s*:\s*"([0-9a-fA-F-]{8,32})"`)
	traceKVRe       = regexp.MustCompile(`(?i)(?:trace[_-]?id|traceId)\s*[=:]\s*["']?([0-9a-fA-F-]{16,64})`)
	spanKVRe        = regexp.MustCompile(`(?i)(?:span[_-]?id|spanId)\s*[=:]\s*["']?([0-9a-fA-F-]{8,32})`)
	correlationRe   = regexp.MustCompile(`(?i)(?:correlation[_-]?id|correlationId|x-correlation-id)\s*[=:]\s*["']?([0-9a-fA-F-]{8,64})`)
	requestIDRe     = regexp.MustCompile(`(?i)(?:request[_-]?id|requestId|x-request-id)\s*[=:]\s*["']?([0-9a-fA-F-]{8,64})`)
	otelTextTraceRe = regexp.MustCompile(`(?i)trace_id[=:\s]+([0-9a-f]{32})`)
	pythonTraceRe   = regexp.MustCompile(`(?i)'?\$traceId'?\s*:\s*['"]([0-9a-fA-F-]{16,64})['"]`)
	uuidInContextRe = regexp.MustCompile(`(?i)(?:trace|request|correlation)[^\n]{0,40}?([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`)
)

// ParseTraceFields extracts distributed tracing identifiers from unstructured log lines.
func ParseTraceFields(message string) map[string]string {
	out := map[string]string{}
	if m := traceParentRe.FindStringSubmatch(message); len(m) >= 3 {
		out["trace_id"] = normalizeTraceID(m[1])
		out["span_id"] = m[2]
	}
	if m := otelTraceJSONRe.FindStringSubmatch(message); len(m) >= 2 && out["trace_id"] == "" {
		out["trace_id"] = normalizeTraceID(m[1])
	}
	if m := otelSpanJSONRe.FindStringSubmatch(message); len(m) >= 2 && out["span_id"] == "" {
		out["span_id"] = m[1]
	}
	if m := traceKVRe.FindStringSubmatch(message); len(m) >= 2 && out["trace_id"] == "" {
		out["trace_id"] = normalizeTraceID(m[1])
	}
	if m := spanKVRe.FindStringSubmatch(message); len(m) >= 2 && out["span_id"] == "" {
		out["span_id"] = m[1]
	}
	if m := otelTextTraceRe.FindStringSubmatch(message); len(m) >= 2 && out["trace_id"] == "" {
		out["trace_id"] = m[1]
	}
	if m := pythonTraceRe.FindStringSubmatch(message); len(m) >= 2 && out["trace_id"] == "" {
		out["trace_id"] = normalizeTraceID(m[1])
	}
	if out["trace_id"] == "" {
		if m := uuidInContextRe.FindStringSubmatch(message); len(m) >= 2 {
			out["trace_id"] = strings.ToLower(m[1])
		}
	}
	if m := correlationRe.FindStringSubmatch(message); len(m) >= 2 {
		out["correlation_id"] = strings.ToLower(m[1])
		if out["trace_id"] == "" {
			out["trace_id"] = out["correlation_id"]
		}
	}
	if m := requestIDRe.FindStringSubmatch(message); len(m) >= 2 {
		out["request_id"] = strings.ToLower(m[1])
		if out["trace_id"] == "" {
			out["trace_id"] = out["request_id"]
		}
	}
	return out
}

func normalizeTraceID(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	raw = strings.ReplaceAll(raw, "-", "")
	if len(raw) == 32 {
		return raw
	}
	if len(raw) == 36 && strings.Count(raw, "-") == 4 {
		return strings.ReplaceAll(raw, "-", "")
	}
	return raw
}

// EnrichLog combines topic classification and trace parsing for hub ingest.
func EnrichLog(message, level string) (topics []string, fields map[string]any) {
	topics, signals := ClassifyLog(message, level)
	fields = map[string]any{}
	for k, v := range signals {
		fields[k] = v
	}
	for k, v := range ParseTraceFields(message) {
		if v != "" {
			fields[k] = v
		}
	}
	if _, ok := fields["trace_id"]; ok {
		topics = appendTopic(topics, "trace")
	}
	fields["topics"] = topics
	return topics, fields
}

func appendTopic(topics []string, t string) []string {
	for _, x := range topics {
		if x == t {
			return topics
		}
	}
	return append(topics, t)
}
