package hub

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

const maxQueryWindow = 24 * time.Hour

var errInvalidRange = errors.New("invalid range")

// requestWindow reads since. A Go duration is an open window ending now.
// start..end, both RFC3339, is a closed window of at most 24 hours.
// until is zero when the window is open.
func requestWindow(r *http.Request, def time.Duration) (from, until time.Time, err error) {
	now := time.Now().UTC()
	from = now.Add(-def)
	raw := strings.TrimSpace(r.URL.Query().Get("since"))
	if raw == "" {
		return from, time.Time{}, nil
	}
	if start, end, ok := splitRange(raw); ok {
		if end.After(now) {
			end = now
		}
		if !end.After(start) || end.Sub(start) > maxQueryWindow {
			return time.Time{}, time.Time{}, errInvalidRange
		}
		return start, end, nil
	}
	d, perr := time.ParseDuration(raw)
	if perr != nil || d <= 0 {
		return from, time.Time{}, nil
	}
	return now.Add(-d), time.Time{}, nil
}

func splitRange(raw string) (time.Time, time.Time, bool) {
	start, end, ok := strings.Cut(raw, "..")
	if !ok || start == "" || end == "" {
		return time.Time{}, time.Time{}, false
	}
	a, err1 := parseTS(start)
	b, err2 := parseTS(end)
	if err1 != nil || err2 != nil {
		return time.Time{}, time.Time{}, false
	}
	return a, b, true
}

func parseTS(raw string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t.UTC(), nil
	}
	t, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}
