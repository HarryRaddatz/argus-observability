import type { SLODefinition } from "@/lib/api"
import type { MeterTone } from "@/lib/meter"
import type { TFunction } from "i18next"

export function budgetTone(pct: number): MeterTone {
  if (pct < 10) return "critical"
  if (pct < 30) return "warning"
  return "default"
}

export function sloObjective(t: TFunction, s: SLODefinition) {
  return s.sli_metric === "latency_p95"
    ? t("slo.latency", { ms: s.latency_threshold_ms, target: s.target })
    : t("slo.availability", { target: s.target })
}

export function sloMetricsLink(s: SLODefinition) {
  const metric = s.sli_metric === "latency_p95" ? "http.duration_ms" : "http.error_rate"
  return `/metrics?mode=compare&metric=${metric}&service=${encodeURIComponent(s.service)}`
}
