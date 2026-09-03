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

export const LOG_TOPICS = [
  { id: "all", label: "Todos" },
  { id: "gc", label: "GC / JVM" },
  { id: "memory", label: "Memória" },
  { id: "oom", label: "OOM" },
  { id: "error", label: "Erros" },
  { id: "performance", label: "Performance" },
] as const

export const LOG_LEVELS = [
  { id: "all", label: "Todos" },
  { id: "error", label: "Error" },
  { id: "warn", label: "Warn" },
  { id: "info", label: "Info" },
  { id: "debug", label: "Debug" },
] as const

export const TIME_RANGES = [
  { id: "15m", label: "15 min" },
  { id: "1h", label: "1 hora" },
  { id: "6h", label: "6 horas" },
  { id: "24h", label: "24 horas" },
] as const

export type SavedView = {
  id: string
  name: string
  metric: string
  containers: string[]
  since: string
  chartType: "area" | "line"
  createdAt: string
}

const VIEWS_KEY = "argus-saved-views"

export function loadSavedViews(): SavedView[] {
  try {
    const raw = localStorage.getItem(VIEWS_KEY)
    if (!raw) return []
    return JSON.parse(raw) as SavedView[]
  } catch {
    return []
  }
}

export function saveView(view: SavedView) {
  const views = loadSavedViews().filter((v) => v.id !== view.id)
  views.unshift(view)
  localStorage.setItem(VIEWS_KEY, JSON.stringify(views.slice(0, 20)))
}

export function deleteView(id: string) {
  const views = loadSavedViews().filter((v) => v.id !== id)
  localStorage.setItem(VIEWS_KEY, JSON.stringify(views))
}
