import { useEffect, useMemo, useState } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"

import { StatToggle } from "@/components/metrics/stat-toggle"
import { MetricMeter } from "@/components/metrics/metric-meter"
import { TimeSeriesChart } from "@/components/metrics/time-series-chart"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import { useQueryPatch, useQueryState } from "@/hooks/use-query-state"
import { parseStatMode, summarize } from "@/lib/stats"
import { fetchMetricSeries, type SeriesPoint, type WorkloadSnapshot } from "@/lib/api"
import { matchesService } from "@/lib/observability"
import { cn } from "@/lib/utils"

type Props = {
  since: string
  workloads: WorkloadSnapshot[]
  members: Set<string> | null
}

type Series = Record<
  "cpu" | "mem" | "memPct" | "httpLatency" | "httpErrors" | "netRx" | "netTx" | "blkRead" | "blkWrite",
  SeriesPoint[]
>

const EMPTY: Series = {
  cpu: [],
  mem: [],
  memPct: [],
  httpLatency: [],
  httpErrors: [],
  netRx: [],
  netTx: [],
  blkRead: [],
  blkWrite: [],
}

const linkClass =
  "text-primary focus-visible:ring-ring inline-flex items-center rounded-sm hover:underline focus-visible:ring-2 focus-visible:outline-none pointer-coarse:min-h-11"

export function ContainerMetrics({ since, workloads, members }: Props) {
  const { t } = useTranslation()
  const [selected] = useQueryState("container", "")
  const [service] = useQueryState("service", "")
  const [statRaw, setStat] = useQueryState("stat", "avg")
  const stat = parseStatMode(statRaw)
  const patch = useQueryPatch()
  const [series, setSeries] = useState<Series>(EMPTY)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [attempt, setAttempt] = useState(0)

  const containers = useMemo(
    () =>
      workloads
        .filter((w) => !members || members.has(w.container))
        .map((w) => w.container)
        .sort(),
    [workloads, members],
  )

  useEffect(() => {
    if (selected || containers.length === 0) return
    const byService = service ? workloads.find((w) => matchesService(w.container, w.service, service)) : undefined
    const busiest = workloads
      .filter((w) => containers.includes(w.container))
      .sort((a, b) => b.cpu_usage - a.cpu_usage || b.memory_usage - a.memory_usage)[0]
    patch({ container: byService?.container ?? busiest?.container ?? containers[0], service: null })
  }, [selected, service, containers, workloads, patch])

  useEffect(() => {
    if (!selected) return
    let cancelled = false
    setLoading(true)
    const get = (metric: string) => fetchMetricSeries(metric, since, selected).then((r) => r.series?.[0]?.points ?? [])
    Promise.all([
      get("cpu.usage"),
      get("memory.usage"),
      get("memory.usage_pct"),
      get("http.duration_ms"),
      get("http.error_rate"),
      get("network.rx"),
      get("network.tx"),
      get("block.read"),
      get("block.write"),
    ])
      .then(([cpu, mem, memPct, httpLatency, httpErrors, netRx, netTx, blkRead, blkWrite]) => {
        if (cancelled) return
        setSeries({ cpu, mem, memPct, httpLatency, httpErrors, netRx, netTx, blkRead, blkWrite })
        setError(null)
      })
      .catch(() => {
        if (!cancelled) setError(t("metrics.loadError"))
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [selected, since, attempt, t])

  const summary = useMemo(() => {
    const pick = (pts: SeriesPoint[]) => {
      const v = summarize(
        pts.map((p) => p.value),
        stat,
      )
      return v === null ? 0 : Math.round(v * 10) / 10
    }
    return { cpu: pick(series.cpu), memPct: pick(series.memPct) }
  }, [series, stat])

  const workload = workloads.find((w) => w.container === selected)
  const statWord = stat === "avg" ? t("metrics.statAvg") : stat === "max" ? t("metrics.statPeak") : stat

  return (
    <div className="flex flex-col gap-6 lg:flex-row">
      <aside className="shrink-0 space-y-2 lg:w-56">
        <h2 className="text-sm font-medium">{t("common.container")}</h2>
        <ScrollArea className="h-[min(420px,55vh)] rounded-md border">
          <ul className="p-1">
            {containers.length === 0 ? (
              <li className="text-muted-foreground p-3 text-sm">{t("metrics.noContainer")}</li>
            ) : (
              containers.map((name) => (
                <li key={name}>
                  <button
                    type="button"
                    aria-current={selected === name ? "true" : undefined}
                    className={cn(
                      "hover:bg-muted focus-visible:ring-ring w-full rounded-md px-3 py-2 text-left text-sm break-all transition-colors focus-visible:ring-2 focus-visible:outline-none pointer-coarse:min-h-11",
                      selected === name && "bg-muted font-medium",
                    )}
                    onClick={() => patch({ container: name })}
                  >
                    {name}
                  </button>
                </li>
              ))
            )}
          </ul>
        </ScrollArea>
      </aside>

      <div className="min-w-0 flex-1 space-y-4">
        {selected ? (
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="space-y-1">
              <h2 className="text-lg font-medium break-all">{selected}</h2>
              {workloads.length > 0 && !containers.includes(selected) ? (
                <p role="status" className="text-muted-foreground text-sm">
                  {t("metrics.notRecent")}
                </p>
              ) : null}
              <div className="flex flex-wrap gap-3 text-sm">
                <Link to={`/logs?container=${encodeURIComponent(selected)}&since=${since}`} className={linkClass}>
                  {t("metrics.logsOf")}
                </Link>
                <Link
                  to={`/traces?service=${encodeURIComponent(workload?.service || selected)}&since=${since}`}
                  className={linkClass}
                >
                  {t("sloExtra.traces")}
                </Link>
                <Link
                  to={`/metrics?mode=compare&containers=${encodeURIComponent(selected)}&since=${since}`}
                  className={linkClass}
                >
                  {t("metrics.compareOthers")}
                </Link>
              </div>
            </div>
            <StatToggle value={stat} onChange={setStat} />
          </div>
        ) : null}

        {error ? (
          <Card className="border-destructive/50" role="alert">
            <CardContent className="flex flex-wrap items-center gap-3 pt-6 text-sm">
              <span className="text-destructive">{error}</span>
              <Button size="sm" variant="outline" onClick={() => setAttempt((n) => n + 1)}>
                {t("common.retry")}
              </Button>
            </CardContent>
          </Card>
        ) : null}

        {!selected ? (
          containers.length === 0 ? null : <Skeleton className="h-64 w-full" />
        ) : (
          <>
            <div className="grid gap-4 lg:grid-cols-2">
              <MetricMeter label={t("metrics.cpuStat", { stat: statWord })} value={summary.cpu} />
              <MetricMeter label={t("metrics.memStat", { stat: statWord })} value={summary.memPct} />
            </div>
            <div className="grid gap-4 lg:grid-cols-2">
              <TimeSeriesChart title={t("containers.cpu")} points={series.cpu} loading={loading} unit="%" />
              <TimeSeriesChart
                title={t("metrics.memoryVsLimit")}
                points={series.memPct}
                loading={loading}
                unit="%"
              />
            </div>
            <div className="grid gap-4 lg:grid-cols-2">
              <TimeSeriesChart
                title={t("metrics.memoryUsed")}
                points={series.mem}
                loading={loading}
                unit=" MiB"
                transform={(v) => Math.round(v / 1024 / 1024)}
              />
              {series.netRx.length > 0 || series.netTx.length > 0 ? (
                <>
                  <TimeSeriesChart title={t("metrics.netRxBytes")} points={series.netRx} loading={loading} unit=" B/s" transform={(v) => Math.round(v)} />
                  <TimeSeriesChart title={t("metrics.netTxBytes")} points={series.netTx} loading={loading} unit=" B/s" transform={(v) => Math.round(v)} />
                </>
              ) : null}
              {series.blkRead.length > 0 || series.blkWrite.length > 0 ? (
                <>
                  <TimeSeriesChart title={t("metrics.diskReadBytes")} points={series.blkRead} loading={loading} unit=" B/s" transform={(v) => Math.round(v)} />
                  <TimeSeriesChart title={t("metrics.diskWriteBytes")} points={series.blkWrite} loading={loading} unit=" B/s" transform={(v) => Math.round(v)} />
                </>
              ) : null}
            </div>
            {series.httpLatency.length > 0 || series.httpErrors.length > 0 ? (
              <div className="grid gap-4 lg:grid-cols-2">
                <TimeSeriesChart
                  title={t("metrics.httpDuration")}
                  description={t("metrics.httpLatencyLogs")}
                  points={series.httpLatency}
                  loading={loading}
                  unit=" ms"
                  transform={(v) => Math.round(v)}
                />
                <TimeSeriesChart
                  title={t("metrics.httpErrorRate")}
                  description={t("metrics.httpErrorStatus")}
                  points={series.httpErrors}
                  loading={loading}
                  unit="%"
                  transform={(v) => Math.round(v * 1000) / 10}
                />
              </div>
            ) : null}
          </>
        )}
      </div>
    </div>
  )
}
