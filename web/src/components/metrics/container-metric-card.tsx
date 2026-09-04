import { Link } from "react-router-dom"

import { MetricMeter } from "@/components/metrics/metric-meter"
import { Sparkline } from "@/components/metrics/time-series-chart"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import type { SeriesPoint, WorkloadSnapshot } from "@/lib/api"
import { formatBytes } from "@/lib/format"

type Props = {
  workload: WorkloadSnapshot
  cpuPoints?: SeriesPoint[]
  memPoints?: SeriesPoint[]
}

export function ContainerMetricCard({ workload, cpuPoints = [], memPoints = [] }: Props) {
  const memPct =
    workload.memory_limit > 0 ? (workload.memory_usage / workload.memory_limit) * 100 : 0

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-2">
        <CardTitle className="truncate text-sm font-medium">{workload.container}</CardTitle>
        <CardDescription>
          {formatBytes(workload.memory_usage)}
          {workload.memory_limit > 0 ? ` / ${formatBytes(workload.memory_limit)}` : ""}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <MetricMeter label="CPU" value={workload.cpu_usage} />
        <MetricMeter label="Memória" value={memPct} />
        <div className="grid grid-cols-2 gap-2">
          <div className="space-y-1">
            <p className="text-muted-foreground text-[10px] uppercase tracking-wide">CPU 1h</p>
            <Sparkline points={cpuPoints} />
          </div>
          <div className="space-y-1">
            <p className="text-muted-foreground text-[10px] uppercase tracking-wide">Mem 1h</p>
            <Sparkline points={memPoints} />
          </div>
        </div>
        <Link
          to={`/metrics?container=${encodeURIComponent(workload.container)}`}
          className="text-primary inline-block text-xs hover:underline"
        >
          Ver detalhes
        </Link>
      </CardContent>
    </Card>
  )
}
