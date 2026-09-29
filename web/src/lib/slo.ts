import type { SLODefinition } from "@/lib/api"
import type { MeterTone } from "@/lib/meter"

export function budgetTone(pct: number): MeterTone {
  if (pct < 10) return "critical"
  if (pct < 30) return "warning"
  return "default"
}

export function sloObjective(s: SLODefinition) {
  return s.sli_metric === "latency_p95"
    ? `p95 abaixo de ${s.latency_threshold_ms} ms em ${s.target}% das requisições`
    : `${s.target}% das requisições sem erro`
}

export function sloMetricsLink(s: SLODefinition) {
  const metric = s.sli_metric === "latency_p95" ? "http.duration_ms" : "http.error_rate"
  return `/metrics?mode=compare&metric=${metric}&service=${encodeURIComponent(s.service)}`
}
