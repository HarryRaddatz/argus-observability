package topology

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/HarryRaddatz/argus-observability/internal/insights"
	"github.com/HarryRaddatz/argus-observability/internal/model"
)

var (
	httpHostRe  = regexp.MustCompile(`(?i)https?://([a-z0-9._-]+)(?::\d+)?`)
	amqpQueueRe = regexp.MustCompile(`(?i)"queue"\s*:\s*"([^"]+)"`)
	amqpChanRe  = regexp.MustCompile(`(?i)"channel"\s*:\s*"([^"]+)"`)
	pathSvcRe   = regexp.MustCompile(`(?i)(?:post|get|put|delete|patch)\s+/([a-z0-9_-]+)`)
)

// TopologyEdge is a directed dependency between services.
type TopologyEdge struct {
	Source string
	Target string
	Kind   string
}

// InferTopologyEdges extracts service-to-service edges from a log line.
func InferTopologyEdges(entry model.LogEntry) []TopologyEdge {
	source := entry.Labels["service"]
	if source == "" {
		source = insights.InferServiceFromContainer(entry.Labels["container"])
	}
	if source == "" {
		return nil
	}

	msg := entry.Message
	var out []TopologyEdge
	seen := map[string]struct{}{}
	add := func(target, kind string) {
		target = sanitizeService(target)
		if target == "" || target == source {
			return
		}
		key := source + "->" + target + ":" + kind
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, TopologyEdge{Source: source, Target: target, Kind: kind})
	}

	if m := httpHostRe.FindStringSubmatch(msg); len(m) >= 2 {
		add(m[1], "http")
	}
	if m := pathSvcRe.FindStringSubmatch(msg); len(m) >= 2 {
		add(m[1], "http")
	}
	if m := amqpQueueRe.FindStringSubmatch(msg); len(m) >= 2 {
		add(queueService(m[1]), "amqp")
	}
	if m := amqpChanRe.FindStringSubmatch(msg); len(m) >= 2 {
		add(queueService(m[1]), "amqp")
	}

	if strings.Contains(strings.ToLower(msg), "rabbitmq") || strings.Contains(strings.ToLower(msg), "message-bus") {
		add("message-bus", "amqp")
	}

	return out
}

func queueService(queue string) string {
	queue = strings.ToLower(queue)
	queue = strings.TrimSuffix(queue, ".queue")
	parts := strings.Split(queue, ".")
	if len(parts) > 0 && parts[0] != "" {
		return sanitizeService(parts[0])
	}
	return sanitizeService(queue)
}

func sanitizeService(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	if idx := strings.Index(name, ":"); idx >= 0 {
		name = name[:idx]
	}
	return name
}

func strField(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}

// InferTopologyFromJSON tries to read url/service targets from structured JSON logs.
func InferTopologyFromJSON(entry model.LogEntry) []TopologyEdge {
	if !strings.HasPrefix(strings.TrimSpace(entry.Message), "{") {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(entry.Message), &m); err != nil {
		return nil
	}
	source := strField(m, "service")
	if source == "" {
		source = entry.Labels["service"]
	}
	if source == "" {
		return nil
	}
	url := strField(m, "url", "path", "target")
	if url == "" {
		return InferTopologyEdges(entry)
	}
	entryCopy := entry
	entryCopy.Message = url
	edges := InferTopologyEdges(entryCopy)
	for i := range edges {
		edges[i].Source = source
	}
	return edges
}
