package docker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

const defaultLogTail = 80

type LogState struct {
	mu       sync.Mutex
	lastSeen map[string]time.Time
}

func newLogState() *LogState {
	return &LogState{lastSeen: map[string]time.Time{}}
}

// NewLogState tracks last-seen log timestamps per container for deduplication.
func NewLogState() *LogState {
	return newLogState()
}

func (s *LogState) shouldKeep(name string, ts time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	prev, ok := s.lastSeen[name]
	if ok && !ts.After(prev) {
		return false
	}
	s.lastSeen[name] = ts
	return true
}

// CollectLogs tails recent stdout/stderr from each filtered container.
func (c *Collector) CollectLogs(ctx context.Context, state *LogState) ([]model.LogEntry, error) {
	if state == nil {
		state = newLogState()
	}
	containers, err := c.listFilteredContainers(ctx)
	if err != nil {
		return nil, err
	}

	tail := defaultLogTail
	var entries []model.LogEntry
	for _, ctr := range containers {
		batch, err := c.containerLogs(ctx, ctr, tail, state)
		if err != nil {
			continue
		}
		entries = append(entries, batch...)
	}
	return entries, nil
}

func (c *Collector) containerLogs(ctx context.Context, ctr containerInfo, tail int, state *LogState) ([]model.LogEntry, error) {
	since := time.Now().UTC().Add(-2 * time.Minute).Unix()
	url := fmt.Sprintf(
		"http://docker/v1.44/containers/%s/logs?stdout=true&stderr=true&timestamps=true&tail=%d&since=%d",
		ctr.ID, tail, since,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("docker logs: %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return nil, err
	}
	lines := coalesceLogLines(demuxDockerLogs(body))
	entityUID, labels := c.entityFor(ctr.Name, ctr.Labels)

	var out []model.LogEntry
	for _, line := range lines {
		if !state.shouldKeep(ctr.Name, line.ts) {
			continue
		}
		out = append(out, model.LogEntry{
			TS:        line.ts,
			Message:   line.message,
			Level:     inferLogLevel(line.message),
			EntityUID: entityUID,
			Labels:    labels,
		})
	}
	return out, nil
}

type parsedLine struct {
	ts      time.Time
	message string
}

func demuxDockerLogs(body []byte) []parsedLine {
	if len(body) == 0 {
		return nil
	}
	if body[0] != 1 && body[0] != 2 {
		return parseLogLines(string(body))
	}
	var chunks []byte
	for len(body) >= 8 {
		size := binary.BigEndian.Uint32(body[4:8])
		if size == 0 || len(body) < 8+int(size) {
			break
		}
		chunks = append(chunks, body[8:8+size]...)
		body = body[8+size:]
	}
	return parseLogLines(string(chunks))
}

func parseLogLines(raw string) []parsedLine {
	var out []parsedLine
	sc := bufio.NewScanner(bytes.NewReader([]byte(raw)))
	sc.Buffer(make([]byte, 64*1024), 256*1024)
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		ts, msg, ok := splitDockerTimestamp(line)
		if !ok {
			// Continuation without docker prefix — attach timestamp zero for coalescing.
			out = append(out, parsedLine{message: strings.TrimRight(line, "\r")})
			continue
		}
		out = append(out, parsedLine{ts: ts, message: msg})
	}
	return out
}

// coalesceLogLines merges physical lines that belong to the same logical log event.
func coalesceLogLines(lines []parsedLine) []parsedLine {
	if len(lines) == 0 {
		return nil
	}
	var out []parsedLine
	var cur *parsedLine

	flush := func() {
		if cur != nil {
			cur.message = strings.TrimSpace(cur.message)
			if cur.message != "" {
				if cur.ts.IsZero() {
					cur.ts = time.Now().UTC()
				}
				out = append(out, *cur)
			}
			cur = nil
		}
	}

	for _, line := range lines {
		line.message = strings.TrimRight(line.message, "\r")
		if strings.TrimSpace(line.message) == "" {
			continue
		}

		if cur == nil {
			cur = &parsedLine{ts: line.ts, message: line.message}
			continue
		}

		if shouldMergeLogLines(cur.message, line.message) {
			cur.message += "\n" + line.message
			continue
		}

		flush()
		cur = &parsedLine{ts: line.ts, message: line.message}
	}
	flush()
	return out
}

func shouldMergeLogLines(prev, next string) bool {
	if isStandaloneLogLine(next) {
		return false
	}
	if !isIncompleteLog(prev) {
		return false
	}
	return isContinuationLine(next)
}

// isStandaloneLogLine detects a full log event that must not be appended to the previous line.
func isStandaloneLogLine(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if strings.HasPrefix(s, "{") && isBalanced(s) {
		return true
	}
	return hasLeadingTimestamp(s)
}

func isIncompleteLog(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if strings.HasSuffix(s, "{") || strings.HasSuffix(s, "[") || strings.HasSuffix(s, ",") {
		return true
	}
	if strings.Contains(s, "{") || strings.Contains(s, "[") {
		return !isBalanced(s)
	}
	return bracketDepth(s) > 0
}

func isContinuationLine(next string) bool {
	nextT := strings.TrimSpace(next)
	if nextT == "}" || nextT == "]" {
		return true
	}
	if strings.HasPrefix(next, " ") || strings.HasPrefix(next, "\t") {
		return true
	}
	if strings.HasPrefix(nextT, "'$") || (strings.HasPrefix(nextT, "'") && strings.Contains(nextT, ":")) {
		return true
	}
	// JSON / Winston object fragment (e.g. "name": "HandlerRegistry",)
	if strings.HasPrefix(nextT, `"`) && strings.Contains(nextT, `":`) && !strings.HasPrefix(nextT, "{") {
		return true
	}
	return false
}

func hasLeadingTimestamp(s string) bool {
	if len(s) >= 20 && s[4] == '-' && s[7] == '-' && s[10] == 'T' {
		end := strings.IndexAny(s, " \t")
		if end > 10 {
			if _, err := time.Parse(time.RFC3339Nano, s[:end]); err == nil {
				return true
			}
			if _, err := time.Parse(time.RFC3339, s[:end]); err == nil {
				return true
			}
		}
	}
	// go-zero / micro access log: YYMMDD/HHMMSS.mmm,
	if len(s) > 15 && s[6] == '/' && s[13] == '.' {
		return true
	}
	return false
}

func bracketDepth(s string) int {
	depth := 0
	inString := false
	escape := false
	var quote byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inString {
			if escape {
				escape = false
				continue
			}
			if c == '\\' {
				escape = true
				continue
			}
			if c == quote {
				inString = false
			}
			continue
		}
		switch c {
		case '"', '\'':
			inString = true
			quote = c
		case '{', '[':
			depth++
		case '}', ']':
			depth--
		}
	}
	return depth
}

func isBalanced(s string) bool {
	return bracketDepth(s) == 0
}

func splitDockerTimestamp(line string) (time.Time, string, bool) {
	if len(line) < 20 || line[10] != 'T' {
		return time.Time{}, line, false
	}
	idx := strings.Index(line, " ")
	if idx <= 0 {
		return time.Time{}, line, false
	}
	ts, err := time.Parse(time.RFC3339Nano, line[:idx])
	if err != nil {
		ts, err = time.Parse(time.RFC3339, line[:idx])
		if err != nil {
			return time.Time{}, line, false
		}
	}
	return ts, strings.TrimSpace(line[idx+1:]), true
}

func inferLogLevel(msg string) string {
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, `"level":"error"`), strings.Contains(lower, `"level": "error"`):
		return "error"
	case strings.Contains(lower, `"level":"warn"`), strings.Contains(lower, `"level":"debug"`):
		if strings.Contains(lower, `"level":"debug"`) {
			return "debug"
		}
		return "warn"
	case strings.Contains(lower, "fatal"), strings.Contains(lower, "panic"), strings.Contains(lower, " error"), strings.HasPrefix(lower, "error"):
		return "error"
	case strings.Contains(lower, "warn"), strings.Contains(lower, "warning"):
		return "warn"
	case strings.Contains(lower, "debug"), strings.Contains(lower, "trace"):
		return "debug"
	default:
		return "info"
	}
}
