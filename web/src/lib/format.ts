import i18n, { localeBcp47 } from "@/i18n"

const cache = new Map<string, Intl.NumberFormat | Intl.DateTimeFormat>()

function cached<T extends Intl.NumberFormat | Intl.DateTimeFormat>(key: string, make: (locale: string) => T): T {
  const locale = localeBcp47(i18n.language)
  const id = `${locale}|${key}`
  let f = cache.get(id)
  if (!f) {
    f = make(locale)
    cache.set(id, f)
  }
  return f as T
}

function decimals(digits: number) {
  return cached(`n${digits}`, (locale) => new Intl.NumberFormat(locale, { minimumFractionDigits: digits, maximumFractionDigits: digits }))
}

export function formatNumber(value: number, digits = 0) {
  return decimals(digits).format(value)
}

export function formatPercent(value: number, digits = 1) {
  return `${formatNumber(value, digits)}%`
}

export function formatBytes(value: number) {
  if (value < 1024) return `${formatNumber(Math.round(value))}\u00a0B`
  if (value < 1024 * 1024) return `${formatNumber(value / 1024, 1)}\u00a0KiB`
  if (value < 1024 * 1024 * 1024) return `${formatNumber(value / 1024 / 1024, 1)}\u00a0MiB`
  return `${formatNumber(value / 1024 / 1024 / 1024, 2)}\u00a0GiB`
}

export function formatDuration(ms: number) {
  if (ms < 1000) return `${formatNumber(ms)}\u00a0ms`
  return `${formatNumber(ms / 1000, 2)}\u00a0s`
}

export function formatTime(ts: string) {
  return cached("hm", (locale) => new Intl.DateTimeFormat(locale, { hour: "2-digit", minute: "2-digit" })).format(new Date(ts))
}

export function formatClock(ts: string) {
  return cached("hms", (locale) => new Intl.DateTimeFormat(locale, { hour: "2-digit", minute: "2-digit", second: "2-digit" })).format(new Date(ts))
}

export function formatDateTime(ts: string) {
  return cached("dt", (locale) => new Intl.DateTimeFormat(locale, { dateStyle: "short", timeStyle: "medium" })).format(new Date(ts))
}

export function chartPoints(points: { ts: string; value: number }[], transform?: (v: number) => number) {
  return points.map((p, i) => ({
    i,
    time: formatTime(p.ts),
    value: transform ? transform(p.value) : p.value,
  }))
}

const ansiEscape = new RegExp(String.raw`\u001b\[[0-9;?]*[ -/]*[@-~]`, "g")

export function stripAnsi(text: string): string {
  return text.includes("\u001b") ? text.replace(ansiEscape, "") : text
}
