import { memo } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"

import { MetricMeter } from "@/components/metrics/metric-meter"
import { Sparkline } from "@/components/metrics/time-series-chart"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import type { ContainerFleetStatus, SeriesPoint, WorkloadSnapshot } from "@/lib/api"
import { formatBytes } from "@/lib/format"
import { stateLabel, stateVariant } from "@/lib/container-state"
import { textTone } from "@/lib/meter"

type Props = {
  workload: WorkloadSnapshot
  status?: ContainerFleetStatus
  cpuPoints?: SeriesPoint[]
  memPoints?: SeriesPoint[]
}

const linkClass =
  "text-primary focus-visible:ring-ring inline-flex items-center rounded-sm hover:underline focus-visible:ring-2 focus-visible:outline-none pointer-coarse:min-h-11"

export const ContainerMetricCard = memo(function ContainerMetricCard({ workload, status, cpuPoints = [], memPoints = [] }: Props) {
  const { t } = useTranslation()
  const memPct =
    workload.memory_limit > 0 ? (workload.memory_usage / workload.memory_limit) * 100 : 0
  const state = status?.state?.toLowerCase()
  const issues: string[] = []
  if (status?.oom_killed) issues.push(t("unstable.oom"))
  if (status?.health === "unhealthy") issues.push(t("unstable.unhealthy"))
  if (status && status.restart_count > 0) issues.push(t("card.restarts", { count: status.restart_count }))

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <div className="flex min-w-0 items-start justify-between gap-2">
          <CardTitle className="min-w-0 truncate text-sm font-medium" title={workload.container}>
            {workload.container}
          </CardTitle>
          {state ? <Badge variant={stateVariant(state)} className="shrink-0">{stateLabel(t, state)}</Badge> : null}
        </div>
        <CardDescription>
          {workload.memory_limit > 0
            ? t("card.memoryOf", { used: formatBytes(workload.memory_usage), limit: formatBytes(workload.memory_limit) })
            : formatBytes(workload.memory_usage)}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <MetricMeter label={t("containers.cpu")} value={workload.cpu_usage} />
        <MetricMeter label={t("containers.memory")} value={memPct} />
        <div className="grid grid-cols-2 gap-2">
          <div className="space-y-1">
            <p className="text-muted-foreground text-xs">{t("card.cpuHour")}</p>
            <Sparkline points={cpuPoints} label={t("containers.cpu")} unit="%" />
          </div>
          <div className="space-y-1">
            <p className="text-muted-foreground text-xs">{t("card.memHour")}</p>
            <Sparkline points={memPoints} label={t("containers.memory")} unit="%" />
          </div>
        </div>
        {issues.length > 0 ? (
          <p className={`text-xs font-medium ${textTone.warning}`}>
            {t("severity.warning")}: {issues.join("; ")}
          </p>
        ) : null}
        <div className="flex gap-4 text-xs">
          <Link to={`/metrics?container=${encodeURIComponent(workload.container)}`} className={linkClass}>
            {t("nav.metrics")}
          </Link>
          <Link to={`/logs?container=${encodeURIComponent(workload.container)}`} className={linkClass}>
            {t("nav.logs")}
          </Link>
        </div>
      </CardContent>
    </Card>
  )
})
