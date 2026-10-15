import type { TFunction } from "i18next"

export const STAT_MODES = ["avg", "p85", "p90", "p95", "max"] as const
export type StatMode = (typeof STAT_MODES)[number]

export function parseStatMode(value: string): StatMode {
  return (STAT_MODES as readonly string[]).includes(value) ? (value as StatMode) : "avg"
}

const PERCENTILE: Partial<Record<StatMode, number>> = { p85: 0.85, p90: 0.9, p95: 0.95 }

// Percentiles interpolate linearly between the two closest ranks.
export function summarize(values: number[], mode: StatMode): number | null {
  if (values.length === 0) return null
  if (mode === "max") return values.reduce((a, b) => (b > a ? b : a), values[0])
  const q = PERCENTILE[mode]
  if (q === undefined) return values.reduce((a, b) => a + b, 0) / values.length
  const sorted = [...values].sort((a, b) => a - b)
  const rank = q * (sorted.length - 1)
  const low = Math.floor(rank)
  const high = Math.ceil(rank)
  return sorted[low] + (sorted[high] - sorted[low]) * (rank - low)
}

export function statLabel(t: TFunction, mode: StatMode): string {
  return t(`chart.stats.${mode}`)
}
