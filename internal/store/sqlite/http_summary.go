package sqlite

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
)

func (s *SQLite) QueryHTTPServiceSummary(ctx context.Context, since time.Time) ([]model.HTTPServiceSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT metric_name, value, labels_json FROM metric_points
WHERE metric_name IN ('http.duration_ms', 'http.requests', 'http.errors') AND ts >= ?
`, since.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type agg struct {
		service   string
		requests  int
		errors    int
		latencies []float64
	}
	byService := map[string]*agg{}

	for rows.Next() {
		var metricName, labelsJSON string
		var value float64
		if err := rows.Scan(&metricName, &value, &labelsJSON); err != nil {
			return nil, err
		}
		var labels model.Labels
		_ = json.Unmarshal([]byte(labelsJSON), &labels)
		svc := labels["service"]
		if svc == "" {
			svc = inferServiceFromContainer(labels["container"])
		}
		if svc == "" {
			continue
		}
		a := byService[svc]
		if a == nil {
			a = &agg{service: svc}
			byService[svc] = a
		}
		switch metricName {
		case "http.requests":
			a.requests += int(value)
		case "http.errors":
			a.errors += int(value)
		case "http.duration_ms":
			a.latencies = append(a.latencies, value)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]model.HTTPServiceSummary, 0, len(byService))
	for _, a := range byService {
		sum := model.HTTPServiceSummary{Service: a.service, Requests: a.requests, Errors: a.errors}
		if a.requests > 0 {
			sum.ErrorRate = float64(a.errors) / float64(a.requests)
		}
		if len(a.latencies) > 0 {
			var total, max float64
			for _, d := range a.latencies {
				total += d
				if d > max {
					max = d
				}
			}
			sum.AvgLatencyMs = total / float64(len(a.latencies))
			sum.MaxLatencyMs = max
		}
		out = append(out, sum)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].AvgLatencyMs > out[j].AvgLatencyMs
	})
	return out, nil
}

func inferServiceFromContainer(container string) string {
	c := container
	if len(c) > 6 && c[:6] == "venuz-" {
		c = c[6:]
	}
	for i := len(c) - 1; i > 0; i-- {
		if c[i] == '-' {
			suffix := c[i+1:]
			allDigits := len(suffix) > 0
			for _, ch := range suffix {
				if ch < '0' || ch > '9' {
					allDigits = false
					break
				}
			}
			if allDigits {
				return c[:i]
			}
		}
	}
	return c
}
