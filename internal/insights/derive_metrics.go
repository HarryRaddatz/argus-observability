package insights

import (
	"github.com/HarryRaddatz/argus-observability/internal/model"
)

// DeriveMetricsFromLog converts structured HTTP exit logs into metric points.
func DeriveMetricsFromLog(entry model.LogEntry) []model.MetricPoint {
	sig, ok := ParseHTTPLog(entry.Message)
	if !ok || !sig.IsExit {
		return nil
	}

	labels := model.Labels{}
	for k, v := range entry.Labels {
		labels[k] = v
	}
	service := sig.Service
	if service == "" {
		service = labels["service"]
	}
	if service == "" {
		service = inferServiceFromContainer(labels["container"])
	}
	if service != "" {
		labels["service"] = service
	}
	if sig.Method != "" {
		labels["method"] = sig.Method
	}

	var out []model.MetricPoint
	out = append(out, point(entry, labels, "http.requests", 1))
	if sig.Status >= 400 {
		out = append(out, point(entry, labels, "http.errors", 1))
	}
	errRate := 0.0
	if sig.Status >= 400 {
		errRate = 1
	}
	out = append(out, point(entry, labels, "http.error_rate", errRate))
	if sig.Status > 0 {
		out = append(out, point(entry, labels, "http.status", float64(sig.Status)))
	}
	if sig.DurationMs > 0 {
		out = append(out, point(entry, labels, "http.duration_ms", sig.DurationMs))
	}
	return out
}

func point(entry model.LogEntry, labels model.Labels, name string, value float64) model.MetricPoint {
	return model.MetricPoint{
		MetricName: name,
		TS:         entry.TS,
		Value:      value,
		EntityUID:  entry.EntityUID,
		Labels:     labels,
	}
}
