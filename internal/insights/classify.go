package insights

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	gcPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\[GC\b`),
		regexp.MustCompile(`(?i)\bGC\s*\(`),
		regexp.MustCompile(`(?i)G1\s+Evacuation\s+Pause`),
		regexp.MustCompile(`(?i)Pause\s+Young`),
		regexp.MustCompile(`(?i)Full\s+GC`),
		regexp.MustCompile(`(?i)Concurrent\s+Mark`),
		regexp.MustCompile(`(?i)CMS\s+Initial\s+Mark`),
		regexp.MustCompile(`(?i)Metaspace`),
		regexp.MustCompile(`(?i)garbage\s+collect`),
	}
	memPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)OutOfMemory`),
		regexp.MustCompile(`(?i)out of memory`),
		regexp.MustCompile(`(?i)MemoryError`),
		regexp.MustCompile(`(?i)heap limit`),
		regexp.MustCompile(`(?i)cannot allocate memory`),
		regexp.MustCompile(`(?i)memory:\s*\d+`),
		regexp.MustCompile(`(?i)rss[:\s]+\d`),
	}
	perfPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)slow query`),
		regexp.MustCompile(`(?i)timeout`),
		regexp.MustCompile(`(?i)deadline exceeded`),
		regexp.MustCompile(`(?i)latency`),
		regexp.MustCompile(`(?i)took \d+(ms|s)`),
	}
)

// ClassifyLog assigns semantic topics and lightweight signals for filtering and insights.
func ClassifyLog(message, level string) (topics []string, signals map[string]any) {
	signals = map[string]any{}
	lower := strings.ToLower(message)

	if matchesAny(message, gcPatterns) {
		topics = append(topics, "gc")
		if pause := extractGCPauseMs(message); pause > 0 {
			signals["gc_pause_ms"] = pause
		}
	}
	if matchesAny(message, memPatterns) {
		topics = append(topics, "memory")
	}
	if level == "error" || strings.Contains(lower, "exception") || strings.Contains(lower, "fatal") {
		topics = append(topics, "error")
	}
	if matchesAny(message, perfPatterns) {
		topics = append(topics, "performance")
	}
	if strings.Contains(lower, "oom") || strings.Contains(lower, "killed") {
		topics = append(topics, "oom")
	}

	if len(topics) == 0 {
		topics = []string{"general"}
	}
	return topics, signals
}

func matchesAny(text string, patterns []*regexp.Regexp) bool {
	for _, p := range patterns {
		if p.MatchString(text) {
			return true
		}
	}
	return false
}

var gcPauseRe = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*(ms|s)\]`)

func extractGCPauseMs(message string) float64 {
	m := gcPauseRe.FindStringSubmatch(message)
	if len(m) < 3 {
		return 0
	}
	val := parseFloat(m[1])
	if m[2] == "s" {
		return val * 1000
	}
	return val
}

func parseFloat(s string) float64 {
	var f float64
	_, _ = fmt.Sscanf(s, "%f", &f)
	return f
}

// TopicQueryPatterns maps UI topic keys to SQL-friendly substrings for log search.
func TopicQueryPatterns(topic string) []string {
	switch topic {
	case "gc":
		return []string{"[GC", "GC(", "G1 Evacuation", "Full GC", "garbage collect"}
	case "memory":
		return []string{"OutOfMemory", "memory:", "heap", "MemoryError", "rss"}
	case "error":
		return []string{"error", "exception", "fatal", "ERR"}
	case "performance":
		return []string{"timeout", "slow query", "latency", "deadline exceeded"}
	case "oom":
		return []string{"oom", "OutOfMemory", "killed", "heap limit"}
	default:
		return nil
	}
}
