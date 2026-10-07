const MAX_MS = 24 * 60 * 60 * 1000

export function isCustomRange(value: string): boolean {
  return value.includes("..")
}

export function splitRange(value: string): { start: string; end: string } | null {
  const i = value.indexOf("..")
  if (i <= 0) return null
  const start = value.slice(0, i)
  const end = value.slice(i + 2)
  if (!start || !end) return null
  return { start, end }
}

export function toLocalInput(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ""
  const pad = (n: number) => String(n).padStart(2, "0")
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function toRFC3339(d: Date): string {
  return d.toISOString().replace(/\.\d{3}Z$/, "Z")
}

// rangeFromLocal returns start..end in UTC, or null when the interval is empty or longer than 24 hours.
export function rangeFromLocal(startLocal: string, endLocal: string): string | null {
  const start = new Date(startLocal)
  const end = new Date(endLocal)
  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) return null
  if (end.getTime() <= start.getTime()) return null
  if (end.getTime() - start.getTime() > MAX_MS) return null
  return `${toRFC3339(start)}..${toRFC3339(end)}`
}

export function defaultLocalRange(): { start: string; end: string } {
  const end = new Date()
  const start = new Date(end.getTime() - 60 * 60 * 1000)
  return { start: toLocalInput(start.toISOString()), end: toLocalInput(end.toISOString()) }
}
