package model

import "time"

type Labels map[string]string

type MetricPoint struct {
	MetricName string    `json:"metric_name"`
	TS         time.Time `json:"ts"`
	Value      float64   `json:"value"`
	EntityUID  string    `json:"entity_uid"`
	Labels     Labels    `json:"labels"`
}

type LogEntry struct {
	TS        time.Time      `json:"ts"`
	Message   string         `json:"message"`
	Level     string         `json:"level"`
	EntityUID string         `json:"entity_uid"`
	Labels    Labels         `json:"labels"`
	Fields    map[string]any `json:"fields,omitempty"`
}

type Event struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	TS        time.Time      `json:"ts"`
	Severity  string         `json:"severity"`
	Source    string         `json:"source"`
	EntityUID string         `json:"entity_uid"`
	Labels    Labels         `json:"labels"`
	Payload   map[string]any `json:"payload,omitempty"`
}

type AgentRegistration struct {
	AgentID  string `json:"agent_id"`
	HostID   string `json:"host_id"`
	Runtime  string `json:"runtime"`
	Labels   Labels `json:"labels"`
	LastSeen time.Time
}

type AgentSession struct {
	SessionID string `json:"session_id"`
	Interval  int    `json:"interval_seconds"`
}

type QuerySeries struct {
	MetricName string       `json:"metric_name"`
	Points     []SeriesPoint `json:"points"`
}

type SeriesPoint struct {
	TS    time.Time `json:"ts"`
	Value float64   `json:"value"`
}
