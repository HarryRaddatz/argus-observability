package insights

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

var (
	uuidRe      = regexp.MustCompile(`(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	hex32Re     = regexp.MustCompile(`(?i)\b[0-9a-f]{32}\b`)
	numberRe    = regexp.MustCompile(`\b\d+(?:\.\d+)?\b`)
	isoTimeRe   = regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?`)
	winstonTsRe = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z`)
)

// NormalizeLogMessage strips volatile fragments for pattern grouping.
func NormalizeLogMessage(message string) string {
	s := strings.TrimSpace(message)
	if s == "" {
		return ""
	}
	s = isoTimeRe.ReplaceAllString(s, "<ts>")
	s = winstonTsRe.ReplaceAllString(s, "<ts>")
	s = uuidRe.ReplaceAllString(s, "<uuid>")
	s = hex32Re.ReplaceAllString(s, "<hex>")
	s = numberRe.ReplaceAllString(s, "<n>")
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 240 {
		s = s[:240]
	}
	return s
}

// PatternKey returns a stable hash for a normalized pattern.
func PatternKey(normalized string) string {
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:16])
}
