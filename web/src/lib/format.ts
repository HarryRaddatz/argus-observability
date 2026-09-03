export function formatPercent(value: number) {
  return `${value.toFixed(1)}%`
}

export function formatBytes(value: number) {
  if (value < 1024) return `${Math.round(value)} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KiB`
  if (value < 1024 * 1024 * 1024) return `${(value / 1024 / 1024).toFixed(1)} MiB`
  return `${(value / 1024 / 1024 / 1024).toFixed(2)} GiB`
}

export function formatTime(ts: string) {
  return new Date(ts).toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" })
}

export function chartPoints(points: { ts: string; value: number }[], transform?: (v: number) => number) {
  return points.map((p, i) => ({
    i,
    time: formatTime(p.ts),
    value: transform ? transform(p.value) : p.value,
  }))
}
