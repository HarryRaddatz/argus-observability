import { useEffect, useMemo, useState } from "react"
import { useSearchParams } from "react-router-dom"

import { TimeSeriesChart } from "@/components/metrics/time-series-chart"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"
import { fetchMetricSeries, listWorkloads } from "@/lib/api"
import { TIME_RANGES } from "@/lib/observability"
import { formatBytes, formatPercent } from "@/lib/format"

export function MetricsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const selected = searchParams.get("container") ?? ""
  const since = searchParams.get("since") ?? "1h"

  const [containers, setContainers] = useState<string[]>([])
  const [cpuPoints, setCpuPoints] = useState<{ ts: string; value: number }[]>([])
  const [memPoints, setMemPoints] = useState<{ ts: string; value: number }[]>([])
  const [memPctPoints, setMemPctPoints] = useState<{ ts: string; value: number }[]>([])
  const [memLimit, setMemLimit] = useState<number>(0)
  const [httpLatency, setHttpLatency] = useState<{ ts: string; value: number }[]>([])
  const [httpErrorRate, setHttpErrorRate] = useState<{ ts: string; value: number }[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    listWorkloads("1h")
      .then((wl) => {
        const names = wl.map((w) => w.container).sort()
        setContainers(names)
        if (!selected && names.length > 0) {
          setSearchParams({ container: names[0], since }, { replace: true })
        }
      })
      .catch(() => setContainers([]))
  }, [selected, since, setSearchParams])

  useEffect(() => {
    if (!selected) {
      setLoading(false)
      return
    }
    let cancelled = false
    setLoading(true)
    Promise.all([
      fetchMetricSeries("cpu.usage", since, selected),
      fetchMetricSeries("memory.usage", since, selected),
      fetchMetricSeries("memory.usage_pct", since, selected),
      fetchMetricSeries("memory.limit", since, selected),
      fetchMetricSeries("http.duration_ms", since, selected),
      fetchMetricSeries("http.error_rate", since, selected),
    ])
      .then(([cpu, mem, memPct, lim, httpLat, httpErr]) => {
        if (cancelled) return
        setCpuPoints(cpu.series?.[0]?.points ?? [])
        setMemPoints(mem.series?.[0]?.points ?? [])
        setMemPctPoints(memPct.series?.[0]?.points ?? [])
        const limits = lim.series?.[0]?.points ?? []
        setMemLimit(limits.length > 0 ? limits[limits.length - 1].value : 0)
        setHttpLatency(httpLat.series?.[0]?.points ?? [])
        setHttpErrorRate(httpErr.series?.[0]?.points ?? [])
        setError(null)
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof Error ? e.message : "Erro ao carregar séries")
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [selected, since])

  const latestCpu = useMemo(
    () => (cpuPoints.length ? cpuPoints[cpuPoints.length - 1].value : 0),
    [cpuPoints],
  )
  const latestMem = useMemo(
    () => (memPoints.length ? memPoints[memPoints.length - 1].value : 0),
    [memPoints],
  )
  const latestMemPct = useMemo(
    () => (memPctPoints.length ? memPctPoints[memPctPoints.length - 1].value : 0),
    [memPctPoints],
  )

  return (
    <div className="flex flex-col gap-6 lg:flex-row">
      <aside className="lg:w-56 shrink-0 space-y-4">
        <div>
          <h2 className="mb-2 text-sm font-medium">Período</h2>
          <div className="flex flex-wrap gap-1">
            {TIME_RANGES.map((r) => (
              <Button
                key={r.id}
                size="sm"
                variant={since === r.id ? "default" : "outline"}
                onClick={() =>
                  setSearchParams({ container: selected, since: r.id }, { replace: true })
                }
              >
                {r.label}
              </Button>
            ))}
          </div>
        </div>
        <div>
          <h2 className="mb-2 text-sm font-medium">Container</h2>
          <ScrollArea className="h-[min(360px,50vh)] rounded-md border">
            <ul className="p-1">
              {containers.length === 0 ? (
                <li className="text-muted-foreground p-3 text-sm">Nenhum workload.</li>
              ) : (
                containers.map((name) => (
                  <li key={name}>
                    <button
                      type="button"
                      className={cn(
                        "hover:bg-muted w-full rounded-md px-3 py-2 text-left text-sm transition-colors",
                        selected === name && "bg-muted font-medium",
                      )}
                      onClick={() => setSearchParams({ container: name, since })}
                    >
                      {name}
                    </button>
                  </li>
                ))
              )}
            </ul>
          </ScrollArea>
        </div>
      </aside>

      <div className="min-w-0 flex-1 space-y-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">{selected || "Métricas"}</h1>
          <p className="text-muted-foreground text-sm">
            CPU, memória absoluta e percentual — útil para detectar pressão antes do OOM.
          </p>
        </div>

        {error ? (
          <Card className="border-destructive/50">
            <CardContent className="text-destructive pt-6 text-sm">{error}</CardContent>
          </Card>
        ) : null}

        {!selected && loading ? (
          <Skeleton className="h-64 w-full" />
        ) : selected ? (
          <>
            <div className="grid gap-4 sm:grid-cols-3">
              <Card>
                <CardContent className="pt-6">
                  <p className="text-muted-foreground text-xs">CPU agora</p>
                  <p className="text-2xl font-semibold tabular-nums">{formatPercent(latestCpu)}</p>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="pt-6">
                  <p className="text-muted-foreground text-xs">Memória agora</p>
                  <p className="text-2xl font-semibold tabular-nums">
                    {formatBytes(latestMem)}
                    {memLimit > 0 ? (
                      <span className="text-muted-foreground text-base font-normal">
                        {" "}
                        / {formatBytes(memLimit)}
                      </span>
                    ) : null}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="pt-6">
                  <p className="text-muted-foreground text-xs">Memória %</p>
                  <p
                    className={cn(
                      "text-2xl font-semibold tabular-nums",
                      latestMemPct >= 90 && "text-destructive",
                      latestMemPct >= 75 && latestMemPct < 90 && "text-amber-600",
                    )}
                  >
                    {formatPercent(latestMemPct)}
                  </p>
                </CardContent>
              </Card>
            </div>
            <TimeSeriesChart
              title="CPU"
              description={`Uso percentual — ${since}`}
              points={cpuPoints}
              loading={loading}
              unit="%"
            />
            <TimeSeriesChart
              title="Memória %"
              description="Pressão de heap/RSS vs limite do container"
              points={memPctPoints}
              loading={loading}
              unit="%"
            />
            <TimeSeriesChart
              title="Memória (MiB)"
              description="Uso absoluto — correlacionar com logs GC"
              points={memPoints}
              loading={loading}
              unit=" MiB"
              transform={(v) => Math.round(v / 1024 / 1024)}
            />
            {httpLatency.length > 0 || httpErrorRate.length > 0 ? (
              <>
                <TimeSeriesChart
                  title="Latência HTTP"
                  description="Derivada de logs event=exit e linhas response Venuz"
                  points={httpLatency}
                  loading={loading}
                  unit=" ms"
                  transform={(v) => Math.round(v)}
                />
                <TimeSeriesChart
                  title="Taxa de erro HTTP"
                  description="Ratio status ≥ 400 por requisição (média no período)"
                  points={httpErrorRate}
                  loading={loading}
                  unit="%"
                  transform={(v) => Math.round(v * 1000) / 10}
                />
              </>
            ) : null}
          </>
        ) : null}
      </div>
    </div>
  )
}
