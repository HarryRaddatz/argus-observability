import { useCallback, useState } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"

import { UsageBar } from "@/components/metrics/metric-meter"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { usePolling } from "@/hooks/use-polling"
import { fetchSLOStatuses, type SLOStatus } from "@/lib/api"
import { budgetTone, sloMetricsLink, sloObjective } from "@/lib/slo"

export function SLOList() {
  const { t } = useTranslation()
  const [rows, setRows] = useState<SLOStatus[]>([])
  const [loading, setLoading] = useState(true)

  const load = useCallback(() => {
    fetchSLOStatuses()
      .then(setRows)
      .catch(() => setRows([]))
      .finally(() => setLoading(false))
  }, [])

  usePolling(load)

  if (loading) return <Skeleton className="h-48 w-full" />
  if (rows.length === 0) {
    return <p className="text-muted-foreground text-sm">{t("slo.empty")}</p>
  }

  const sorted = [...rows].sort((a, b) => a.error_budget_remaining - b.error_budget_remaining)

  return (
    <div className="grid gap-4 md:grid-cols-2">
      {sorted.map((row) => (
        <Card key={row.slo.id} className={row.breached ? "border-destructive/40" : undefined}>
          <CardHeader className="pb-2">
            <div className="flex items-start justify-between gap-2">
              <CardTitle className="text-base font-medium">{row.slo.name}</CardTitle>
              <Badge variant={row.breached ? "destructive" : "outline"}>{row.breached ? t("overview.violated") : t("sloExtra.ok")}</Badge>
            </div>
            <p className="text-sm">{sloObjective(t, row.slo)}</p>
            <p className="text-muted-foreground text-xs">
              {t("sloExtra.window", { service: row.slo.service, hours: row.slo.window_hours })}
            </p>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div className="space-y-1">
              <div className="flex justify-between text-xs">
                <span>{t("sloExtra.compliance")}</span>
                <span className="font-mono tabular-nums">{row.compliance.toFixed(2)}%</span>
              </div>
              <UsageBar value={row.compliance} tone={row.breached ? "critical" : "default"} label={t("sloExtra.complianceOf", { name: row.slo.name })} />
              <p className="text-muted-foreground text-xs tabular-nums">
                {t("sloExtra.requests", { good: row.good_events, total: row.total_events })}
                {row.p95_latency_ms > 0 ? t("sloExtra.p95", { ms: row.p95_latency_ms.toFixed(0) }) : ""}
              </p>
            </div>
            <div className="space-y-1">
              <div className="flex justify-between text-xs">
                <span>{t("sloExtra.budget")}</span>
                <span className="font-mono tabular-nums">{row.error_budget_remaining.toFixed(1)}%</span>
              </div>
              <UsageBar
                value={row.error_budget_remaining}
                tone={budgetTone(row.error_budget_remaining)}
                label={t("sloExtra.budgetOf", { name: row.slo.name })}
              />
            </div>
            {row.slo.service ? (
              <div className="flex flex-wrap gap-3 text-xs">
                <Link className="text-primary hover:underline" to={sloMetricsLink(row.slo)}>
                  {t("slo.httpMetrics")}
                </Link>
                <Link className="text-primary hover:underline" to={`/traces?service=${encodeURIComponent(row.slo.service)}`}>
                  {t("sloExtra.traces")}
                </Link>
              </div>
            ) : null}
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
