const base = import.meta.env.VITE_API_BASE ?? ""

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${base}${path}`, init)
  if (!res.ok) {
    throw new Error(`${res.status} ${res.statusText}`)
  }
  const data = (await res.json()) as T
  return data
}

export function asArray<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : []
}

export type Health = { status: string }

export type SeriesPoint = { ts: string; value: number }

export type ContainerSeries = {
  container: string
  entity_uid: string
  points: SeriesPoint[]
}

export type MetricSeriesResponse = {
  metric_name: string
  series: ContainerSeries[]
}

export type WorkloadSnapshot = {
  container: string
  entity_uid: string
  stack?: string
  service?: string
  labels?: Record<string, string>
  cpu_usage: number
  memory_usage: number
  memory_limit: number
  updated_at: string
}

export type WorkloadGroup = {
  id: string
  name: string
  kind: "stack" | "service" | "custom"
  description?: string
  label_key?: string
  label_value?: string
  containers?: string[]
  member_count?: number
  created_at?: string
  updated_at?: string
}

export type WorkloadGroupInput = {
  name: string
  kind: "stack" | "service" | "custom"
  description?: string
  label_key?: string
  label_value?: string
  containers?: string[]
}

export type WorkloadGroupSummary = {
  group: WorkloadGroup
  member_count: number
  avg_cpu: number
  avg_memory_pct: number
  total_memory: number
  members: WorkloadSnapshot[]
}

export type EventRow = {
  id: string
  type: string
  ts: string
  severity: string
  source: string
  entity_uid: string
  labels: Record<string, string>
  payload?: Record<string, unknown>
}

export type LogRow = {
  ts: string
  message: string
  level: string
  entity_uid: string
  labels: Record<string, string>
  fields?: Record<string, unknown>
}

export type LogSearchParams = {
  q?: string
  since?: string
  level?: string
  container?: string
  topic?: string
  group?: string
  trace_id?: string
}

export type MetricCatalogEntry = {
  name: string
  label: string
  unit: string
}

export type Insight = {
  id: string
  theme: string
  severity: string
  title: string
  summary: string
  container: string
  entity_uid: string
  evidence: Record<string, unknown>
  recommendations: string[]
}

export type InsightsResponse = {
  since: string
  insights: Insight[]
}

export type ContainerFleetStatus = {
  container: string
  entity_uid: string
  service?: string
  state: string
  health?: string
  restart_count: number
  exit_code?: number
  oom_killed?: boolean
  status_text?: string
  updated_at?: string
}

export type FleetSummary = {
  running: number
  exited: number
  restarting: number
  unhealthy: number
  dead: number
  total_restart_count: number
  replicas_up: number
  replicas_total: number
}

export type ServiceReplicaStatus = {
  service: string
  replicas_up: number
  replicas_total: number
  unhealthy: number
  restarting: number
}

export type FleetEventStats = {
  restarts_24h: number
  failures_24h: number
  oom_24h: number
  disconnect_24h: number
}

export type FleetStatusResponse = {
  updated_at: string
  summary: FleetSummary
  services: ServiceReplicaStatus[]
  containers: ContainerFleetStatus[]
  events_24h: FleetEventStats
}

export function getHealth() {
  return request<Health>("/health")
}

export function fetchMetricSeries(metric: string, since = "1h", container?: string, group?: string) {
  const q = new URLSearchParams({ metric, since })
  if (container) q.set("container", container)
  if (group) q.set("group", group)
  return request<MetricSeriesResponse>(`/api/v1/metrics/series?${q}`)
}

export function listWorkloads(since = "30m") {
  return request<WorkloadSnapshot[]>(`/api/v1/workloads?since=${encodeURIComponent(since)}`).then(asArray)
}

export function listEvents(entityUID?: string, since = "24h") {
  const q = new URLSearchParams({ since })
  if (entityUID) q.set("entity_uid", entityUID)
  return request<EventRow[]>(`/api/v1/events?${q}`).then(asArray)
}

export function searchLogs(params: LogSearchParams = {}) {
  const p = new URLSearchParams()
  p.set("since", params.since ?? "1h")
  if (params.q) p.set("q", params.q)
  if (params.level && params.level !== "all") p.set("level", params.level)
  if (params.container && params.container !== "all") p.set("container", params.container)
  if (params.topic && params.topic !== "all") p.set("topic", params.topic)
  if (params.group) p.set("group", params.group)
  if (params.trace_id) p.set("trace_id", params.trace_id)
  return request<LogRow[]>(`/api/v1/logs/search?${p}`).then(asArray)
}

export function listWorkloadGroups() {
  return request<WorkloadGroup[]>(`/api/v1/workload-groups`).then(asArray)
}

export function discoverWorkloadGroups() {
  return request<WorkloadGroup[]>(`/api/v1/workload-groups/discover`).then(asArray)
}

export function getWorkloadGroupSummary(id: string, since = "30m") {
  return request<WorkloadGroupSummary>(
    `/api/v1/workload-groups/${encodeURIComponent(id)}/summary?since=${encodeURIComponent(since)}`,
  )
}

export function createWorkloadGroup(input: WorkloadGroupInput) {
  return request<WorkloadGroup>(`/api/v1/workload-groups`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
}

export function deleteWorkloadGroup(id: string) {
  return fetch(`${base}/api/v1/workload-groups/${encodeURIComponent(id)}`, { method: "DELETE" }).then(
    (res) => {
      if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
    },
  )
}

export function fetchInsights(since = "1h") {
  return request<InsightsResponse>(`/api/v1/insights?since=${encodeURIComponent(since)}`)
}

export function fetchFleetStatus() {
  return request<FleetStatusResponse>(`/api/v1/fleet/status`)
}

export function fetchMetricCatalog() {
  return request<MetricCatalogEntry[]>(`/api/v1/metrics/catalog`).then(asArray)
}

export type HTTPServiceSummary = {
  service: string
  requests: number
  errors: number
  error_rate: number
  avg_latency_ms: number
  max_latency_ms: number
}

export function fetchHTTPSummary(since = "1h") {
  return request<HTTPServiceSummary[]>(
    `/api/v1/metrics/http/summary?since=${encodeURIComponent(since)}`,
  ).then(asArray)
}
