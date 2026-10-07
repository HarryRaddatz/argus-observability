import { formatTime } from "@/lib/format"

export type StackPoint = { ts: string; value: number }

export type StackSeries = { container: string; points?: StackPoint[] | null }

// One row per minute. Several scrapes share a minute; the last one was a
// spike and was being drawn as the whole minute. The row is the mean, and a
// container with no sample that minute is 0 so the stack height is the sum.
export function mergeStacked(
  series: StackSeries[],
  transform?: (v: number) => number,
): Array<Record<string, string | number>> {
  const buckets = new Map<number, Map<string, { sum: number; n: number }>>()
  for (const s of series) {
    for (const p of s.points ?? []) {
      const minute = Math.floor(new Date(p.ts).getTime() / 60_000) * 60_000
      if (!Number.isFinite(minute)) continue
      const value = transform ? transform(p.value) : p.value
      if (!Number.isFinite(value)) continue
      let row = buckets.get(minute)
      if (!row) {
        row = new Map()
        buckets.set(minute, row)
      }
      const cell = row.get(s.container) ?? { sum: 0, n: 0 }
      cell.sum += value
      cell.n += 1
      row.set(s.container, cell)
    }
  }
  const names = series.map((s) => s.container)
  return [...buckets.keys()].sort((a, b) => a - b).map((minute) => {
    const cells = buckets.get(minute)!
    const out: Record<string, string | number> = {
      time: formatTime(new Date(minute).toISOString()),
    }
    for (const name of names) {
      const cell = cells.get(name)
      out[name] = cell && cell.n > 0 ? cell.sum / cell.n : 0
    }
    return out
  })
}
