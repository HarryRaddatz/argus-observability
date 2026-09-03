package insights

import (
	"fmt"
	"strings"
)

// TraceSearchPatterns builds SQL LIKE patterns to match a trace id in fields_json or message.
func TraceSearchPatterns(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	add := func(fragment string) {
		fragment = strings.TrimSpace(fragment)
		if fragment == "" {
			return
		}
		pat := fmt.Sprintf("%%%s%%", fragment)
		if _, ok := seen[pat]; ok {
			return
		}
		seen[pat] = struct{}{}
		out = append(out, pat)
	}

	add(raw)
	add(strings.ToLower(raw))
	norm := normalizeTraceID(raw)
	add(norm)
	if len(norm) == 32 {
		add(formatUUID(norm))
		add(fmt.Sprintf(`"traceId":"%s"`, formatUUID(norm)))
		add(fmt.Sprintf(`"traceId":"%s"`, norm))
		add(fmt.Sprintf(`"trace_id":"%s"`, norm))
	}
	return out
}

func formatUUID(hex32 string) string {
	if len(hex32) != 32 {
		return hex32
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex32[0:8], hex32[8:12], hex32[12:16], hex32[16:20], hex32[20:32])
}
