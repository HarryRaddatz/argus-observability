export const LOG_TOPICS = [
  { id: "all" },
  { id: "gc" },
  { id: "memory" },
  { id: "oom" },
  { id: "error" },
  { id: "performance" },
  { id: "trace" },
] as const

export const LOG_LEVELS = [
  { id: "all" },
  { id: "error" },
  { id: "warn" },
  { id: "info" },
  { id: "debug" },
] as const

export const TIME_RANGES = [
  { id: "15m" },
  { id: "1h" },
  { id: "6h" },
  { id: "24h" },
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

export function matchesService(container: string, service: string | undefined, target: string) {
  if (!target) return true
  return service === target || container === target || container.includes(`-${target}-`) || container.startsWith(`${target}-`)
}
