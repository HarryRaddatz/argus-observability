import type { TFunction } from "i18next"

const RULE_KEYS: Record<string, string> = {
  "cpu-high": "alerts.cpuHigh",
  "memory-high": "alerts.memoryHigh",
}

const METRIC_KEYS: Record<string, string> = {
  "cpu.usage": "metrics.cpuUsage",
  "memory.usage": "metrics.memoryUsage",
  "memory.usage_pct": "metrics.memoryPct",
  "memory.limit": "metrics.memoryLimit",
  "http.duration_ms": "metrics.httpDuration",
  "http.error_rate": "metrics.httpErrorRate",
  "http.requests": "metrics.httpRequests",
  "http.status": "metrics.httpStatus",
  "network.rx": "metrics.netRx",
  "network.tx": "metrics.netTx",
  "block.read": "metrics.diskRead",
  "block.write": "metrics.diskWrite",
}

export function localizeAlertTitle(t: TFunction, ruleId: string, fallback: string): string {
  const key = RULE_KEYS[ruleId]
  return key ? t(key) : fallback
}

export function localizeMetricLabel(t: TFunction, metricName: string, fallback: string): string {
  const key = METRIC_KEYS[metricName]
  return key ? t(key) : fallback
}

export function localizeInsightTheme(t: TFunction, theme: string): string {
  const key = `insightTheme.${theme}`
  const label = t(key)
  return label === key ? theme : label
}
