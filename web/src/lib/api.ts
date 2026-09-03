const base = import.meta.env.VITE_API_BASE ?? ""

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${base}${path}`, init)
  if (!res.ok) {
    throw new Error(`${res.status} ${res.statusText}`)
  }
  return res.json() as Promise<T>
}

export type Health = { status: string }

export type SeriesPoint = { ts: string; value: number }
export type QuerySeries = { metric_name: string; points: SeriesPoint[] }

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
}

export function getHealth() {
  return request<Health>("/health")
}

export function queryMetrics(metric: string, since = "1h") {
  return request<QuerySeries>(`/api/v1/query?metric=${encodeURIComponent(metric)}&since=${encodeURIComponent(since)}`)
}

export function listEvents(entityUID?: string, since = "24h") {
  const q = new URLSearchParams({ since })
  if (entityUID) q.set("entity_uid", entityUID)
  return request<EventRow[]>(`/api/v1/events?${q}`)
}

export function searchLogs(q: string, since = "15m") {
  const params = new URLSearchParams({ since })
  if (q) params.set("q", q)
  return request<LogRow[]>(`/api/v1/logs/search?${params}`)
}
