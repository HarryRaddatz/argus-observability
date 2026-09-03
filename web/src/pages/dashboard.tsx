import { useEffect, useMemo, useState, type ReactNode } from "react"
import { Link } from "react-router-dom"

import { Sparkline } from "@/components/metrics/time-series-chart"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"
import { fetchMetricSeries, fetchHTTPSummary, getHealth, listEvents, listWorkloads, type HTTPServiceSummary } from "@/lib/api"
import { formatBytes, formatPercent } from "@/lib/format"

export function DashboardPage() {
  const [health, setHealth] = useState<"ok" | "error" | "loading">("loading")
  const [events, setEvents] = useState<number | null>(null)
  const [workloads, setWorkloads] = useState<Awaited<ReturnType<typeof listWorkloads>>>([])
  const [cpuSeries, setCpuSeries] = useState<Record<string, { ts: string; value: number }[]>>({})
  const [httpServices, setHttpServices] = useState<HTTPServiceSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        await getHealth()
        if (!cancelled) setHealth("ok")
        const [ev, wl, series, http] = await Promise.all([
          listEvents(undefined, "24h"),
          listWorkloads("30m"),
          fetchMetricSeries("cpu.usage", "1h"),
          fetchHTTPSummary("1h"),
        ])
        if (cancelled) return
        setEvents(ev.length)
        setWorkloads(wl)
        setHttpServices(http.filter((s) => s.requests > 0).slice(0, 8))
        const map: Record<string, { ts: string; value: number }[]> = {}
        for (const s of series.series ?? []) {
          map[s.container] = s.points ?? []
        }
        setCpuSeries(map)
      } catch (e) {
        if (!cancelled) {
          setHealth("error")
          setError(e instanceof Error ? e.message : "Falha ao carregar")
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    load()
    const t = setInterval(load, 30_000)
    return () => {
      cancelled = true
      clearInterval(t)
    }
  }, [])

  const topWorkloads = useMemo(
    () => [...workloads].sort((a, b) => b.cpu_usage - a.cpu_usage).slice(0, 12),
    [workloads],
  )

  const avgCpu = useMemo(() => {
    if (workloads.length === 0) return 0
    return workloads.reduce((a, w) => a + w.cpu_usage, 0) / workloads.length
  }, [workloads])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Visão geral</h1>
        <p className="text-muted-foreground text-sm">Stack Venuz — métricas em tempo quase real.</p>
      </div>

      {error ? (
        <Card className="border-destructive/50">
          <CardContent className="text-destructive pt-6 text-sm">{error}</CardContent>
        </Card>
      ) : null}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard title="Hub" loading={loading}>
          <Badge variant={health === "ok" ? "default" : "destructive"}>
            {health === "loading" ? "…" : health === "ok" ? "Online" : "Offline"}
          </Badge>
        </StatCard>
        <StatCard title="Workloads" description="Containers venuz-*" loading={loading}>
          <p className="text-3xl font-semibold tabular-nums">{workloads.length}</p>
        </StatCard>
        <StatCard title="CPU média" description="Última amostra" loading={loading}>
          <p className="text-3xl font-semibold tabular-nums">{formatPercent(avgCpu)}</p>
        </StatCard>
        <StatCard title="Eventos" description="24h" loading={loading}>
          <p className="text-3xl font-semibold tabular-nums">{events ?? "—"}</p>
        </StatCard>
      </div>

      {httpServices.length > 0 ? (
        <div>
          <div className="mb-3 flex items-center justify-between">
            <h2 className="text-lg font-medium">HTTP por serviço</h2>
            <Link to="/explorer?metric=http.duration_ms" className="text-primary text-sm hover:underline">
              Explorer
            </Link>
          </div>
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            {httpServices.map((s) => (
              <Card key={s.service}>
                <CardHeader className="pb-2">
                  <CardTitle className="truncate text-sm font-medium">{s.service}</CardTitle>
                  <CardDescription>{s.requests} req · 1h</CardDescription>
                </CardHeader>
                <CardContent className="space-y-1 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Latência média</span>
                    <span className="font-mono tabular-nums">{Math.round(s.avg_latency_ms)} ms</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Taxa de erro</span>
                    <span
                      className={cn(
                        "font-mono tabular-nums",
                        s.error_rate >= 0.05 && "text-destructive",
                        s.error_rate >= 0.01 && s.error_rate < 0.05 && "text-amber-600",
                      )}
                    >
                      {formatPercent(s.error_rate * 100)}
                    </span>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      ) : null}

      <div>
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-lg font-medium">Containers</h2>
          <Link to="/metrics" className="text-primary text-sm hover:underline">
            Ver detalhes
          </Link>
        </div>
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {loading
            ? Array.from({ length: 6 }).map((_, i) => (
                <Card key={i}>
                  <CardHeader>
                    <Skeleton className="h-5 w-32" />
                  </CardHeader>
                  <CardContent>
                    <Skeleton className="h-12 w-full" />
                  </CardContent>
                </Card>
              ))
            : topWorkloads.map((w) => (
                <Card key={w.entity_uid}>
                  <CardHeader className="pb-2">
                    <CardTitle className="truncate text-sm font-medium">{w.container}</CardTitle>
                    <CardDescription className="flex justify-between gap-2">
                      <span>CPU {formatPercent(w.cpu_usage)}</span>
                      <span>{formatBytes(w.memory_usage)}</span>
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-2">
                    <Sparkline points={cpuSeries[w.container] ?? []} />
                    <MemoryBar usage={w.memory_usage} limit={w.memory_limit} />
                  </CardContent>
                </Card>
              ))}
        </div>
      </div>
    </div>
  )
}

function StatCard({
  title,
  description,
  loading,
  children,
}: {
  title: string
  description?: string
  loading?: boolean
  children: ReactNode
}) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        {description ? <CardDescription>{description}</CardDescription> : null}
      </CardHeader>
      <CardContent>{loading ? <Skeleton className="h-8 w-16" /> : children}</CardContent>
    </Card>
  )
}

function MemoryBar({ usage, limit }: { usage: number; limit: number }) {
  const pct = limit > 0 ? Math.min(100, (usage / limit) * 100) : 0
  return (
    <div className="space-y-1">
      <div className="text-muted-foreground flex justify-between text-xs">
        <span>RAM</span>
        <span>
          {formatBytes(usage)}
          {limit > 0 ? ` / ${formatBytes(limit)}` : ""}
        </span>
      </div>
      <div className="bg-muted h-1.5 overflow-hidden rounded-full">
        <div
          className="bg-primary h-full transition-all"
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  )
}
