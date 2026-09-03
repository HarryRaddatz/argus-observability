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

type ContainerSeries struct {
	Container string        `json:"container"`
	EntityUID string        `json:"entity_uid"`
	Points    []SeriesPoint `json:"points"`
}

type MetricSeriesResponse struct {
	MetricName string            `json:"metric_name"`
	Series     []ContainerSeries `json:"series"`
}

type WorkloadSnapshot struct {
	Container   string    `json:"container"`
	EntityUID   string    `json:"entity_uid"`
	Stack       string    `json:"stack,omitempty"`
	Service     string    `json:"service,omitempty"`
	Labels      Labels    `json:"labels,omitempty"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	MemoryLimit float64   `json:"memory_limit"`
	UpdatedAt   time.Time `json:"updated_at"`
}

const (
	GroupKindStack   = "stack"
	GroupKindService = "service"
	GroupKindCustom  = "custom"
)

type WorkloadGroup struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Kind        string    `json:"kind"`
	Description string    `json:"description,omitempty"`
	LabelKey    string    `json:"label_key,omitempty"`
	LabelValue  string    `json:"label_value,omitempty"`
	Containers  []string  `json:"containers,omitempty"`
	MemberCount int       `json:"member_count,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WorkloadGroupInput struct {
	Name        string   `json:"name"`
	Kind        string   `json:"kind"`
	Description string   `json:"description,omitempty"`
	LabelKey    string   `json:"label_key,omitempty"`
	LabelValue  string   `json:"label_value,omitempty"`
	Containers  []string `json:"containers,omitempty"`
}

type WorkloadGroupSummary struct {
	Group         WorkloadGroup      `json:"group"`
	MemberCount   int                `json:"member_count"`
	AvgCPU        float64            `json:"avg_cpu"`
	AvgMemoryPct  float64            `json:"avg_memory_pct"`
	TotalMemory   float64            `json:"total_memory"`
	Members       []WorkloadSnapshot `json:"members"`
}

type LogSearchFilter struct {
	Query      string
	EntityUID  string
	Container  string
	Containers []string
	Level      string
	Topic      string
	TraceID    string
	Since      time.Time
	Limit      int
}

type LogTopicCount struct {
	Container string `json:"container"`
	EntityUID string `json:"entity_uid"`
	Topic     string `json:"topic"`
	Count     int    `json:"count"`
}

type InsightsResponse struct {
	Since    string    `json:"since"`
	Insights []Insight `json:"insights"`
}

type Insight struct {
	ID              string         `json:"id"`
	Theme           string         `json:"theme"`
	Severity        string         `json:"severity"`
	Title           string         `json:"title"`
	Summary         string         `json:"summary"`
	Container       string         `json:"container"`
	EntityUID       string         `json:"entity_uid"`
	Evidence        map[string]any `json:"evidence"`
	Recommendations []string       `json:"recommendations"`
}

type ContainerFleetStatus struct {
	Container    string    `json:"container"`
	EntityUID    string    `json:"entity_uid"`
	Service      string    `json:"service"`
	State        string    `json:"state"`
	Health       string    `json:"health,omitempty"`
	RestartCount int       `json:"restart_count"`
	ExitCode     int       `json:"exit_code,omitempty"`
	OOMKilled    bool      `json:"oom_killed,omitempty"`
	StatusText   string    `json:"status_text,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ServiceReplicaStatus struct {
	Service       string `json:"service"`
	ReplicasUp    int    `json:"replicas_up"`
	ReplicasTotal int    `json:"replicas_total"`
	Unhealthy     int    `json:"unhealthy"`
	Restarting    int    `json:"restarting"`
}

type FleetEventStats struct {
	Restarts24h  int `json:"restarts_24h"`
	Failures24h  int `json:"failures_24h"`
	OOM24h       int `json:"oom_24h"`
	Disconnect24h int `json:"disconnect_24h"`
}

type FleetSummary struct {
	Running           int `json:"running"`
	Exited            int `json:"exited"`
	Restarting        int `json:"restarting"`
	Unhealthy         int `json:"unhealthy"`
	Dead              int `json:"dead"`
	TotalRestartCount int `json:"total_restart_count"`
	ReplicasUp        int `json:"replicas_up"`
	ReplicasTotal     int `json:"replicas_total"`
}

type FleetStatusResponse struct {
	UpdatedAt  time.Time              `json:"updated_at"`
	Summary    FleetSummary           `json:"summary"`
	Services   []ServiceReplicaStatus `json:"services"`
	Containers []ContainerFleetStatus `json:"containers"`
	Events24h  FleetEventStats        `json:"events_24h"`
}
