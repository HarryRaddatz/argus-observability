import { Link } from "react-router-dom"

import { MetricMeter } from "@/components/metrics/metric-meter"
import { Sparkline } from "@/components/metrics/time-series-chart"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import type { ContainerFleetStatus, SeriesPoint, WorkloadSnapshot } from "@/lib/api"
import { formatBytes } from "@/lib/format"
import { stateLabel, stateVariant } from "@/lib/container-state"

type Props = {
  workload: WorkloadSnapshot
  status?: ContainerFleetStatus
  cpuPoints?: SeriesPoint[]
  memPoints?: SeriesPoint[]
}

export function ContainerMetricCard({ workload, status, cpuPoints = [], memPoints = [] }: Props) {
  const memPct =
    workload.memory_limit > 0 ? (workload.memory_usage / workload.memory_limit) * 100 : 0
  const state = status?.state?.toLowerCase()

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between gap-2">
          <CardTitle className="truncate text-sm font-medium">{workload.container}</CardTitle>
          {state ? <Badge variant={stateVariant(state)}>{stateLabel(state)}</Badge> : null}
        </div>
        <CardDescription>
          {formatBytes(workload.memory_usage)}
          {workload.memory_limit > 0 ? ` de ${formatBytes(workload.memory_limit)}` : ""}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <MetricMeter label="CPU" value={workload.cpu_usage} />
        <MetricMeter label="Memória" value={memPct} />
        <div className="grid grid-cols-2 gap-2">
          <div className="space-y-1">
            <p className="text-muted-foreground text-xs">CPU na última hora</p>
            <Sparkline points={cpuPoints} />
          </div>
          <div className="space-y-1">
            <p className="text-muted-foreground text-xs">Memória na última hora</p>
            <Sparkline points={memPoints} />
          </div>
        </div>
        {status && (status.restart_count > 0 || status.oom_killed || status.health === "unhealthy") ? (
          <p className="text-amber-600 text-xs font-medium">
            {status.oom_killed ? "Encerrado por falta de memória. " : ""}
            {status.health === "unhealthy" ? "Healthcheck falhando. " : ""}
            {status.restart_count > 0
              ? `${status.restart_count} ${status.restart_count === 1 ? "reinício" : "reinícios"}`
              : ""}
          </p>
        ) : null}
        <div className="flex gap-3 text-xs">
          <Link to={`/metrics?container=${encodeURIComponent(workload.container)}`} className="text-primary hover:underline">
            Métricas
          </Link>
          <Link to={`/logs?container=${encodeURIComponent(workload.container)}`} className="text-primary hover:underline">
            Logs
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}
