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
  cpu_usage: number
  memory_usage: number
  memory_limit: number
  updated_at: string
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

export function getHealth() {
  return request<Health>("/health")
}

export function fetchMetricSeries(metric: string, since = "1h", container?: string) {
  const q = new URLSearchParams({ metric, since })
  if (container) q.set("container", container)
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
  return request<LogRow[]>(`/api/v1/logs/search?${p}`).then(asArray)
}

export function fetchInsights(since = "1h") {
  return request<InsightsResponse>(`/api/v1/insights?since=${encodeURIComponent(since)}`)
}

export function fetchMetricCatalog() {
  return request<MetricCatalogEntry[]>(`/api/v1/metrics/catalog`).then(asArray)
}
