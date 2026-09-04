import { useEffect, useMemo, useState } from "react"
import { Link } from "react-router-dom"

import { ContainerMetricCard } from "@/components/metrics/container-metric-card"
import { StackedAreaChart } from "@/components/metrics/stacked-area-chart"
import { PageHeader } from "@/components/layout/page-header"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { fetchMetricSeries, listWorkloads, type WorkloadSnapshot } from "@/lib/api"
import { formatBytes, formatPercent } from "@/lib/format"

const GRID_LIMIT = 40

export function WorkloadsPage() {
  const [rows, setRows] = useState<WorkloadSnapshot[]>([])
  const [cpuSeries, setCpuSeries] = useState<Record<string, { ts: string; value: number }[]>>({})
  const [memSeries, setMemSeries] = useState<Record<string, { ts: string; value: number }[]>>({})
  const [stackedCpu, setStackedCpu] = useState<Awaited<ReturnType<typeof fetchMetricSeries>>["series"]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [view, setView] = useState<"grid" | "table">("grid")

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const [wl, cpu, memPct] = await Promise.all([
          listWorkloads("30m"),
          fetchMetricSeries("cpu.usage", "1h"),
          fetchMetricSeries("memory.usage_pct", "1h"),
        ])
        if (cancelled) return
        const sorted = [...wl].sort((a, b) => b.cpu_usage - a.cpu_usage)
        setRows(sorted)
        const cpuMap: Record<string, { ts: string; value: number }[]> = {}
        for (const s of cpu.series ?? []) cpuMap[s.container] = s.points ?? []
        setCpuSeries(cpuMap)
        const memMap: Record<string, { ts: string; value: number }[]> = {}
        for (const s of memPct.series ?? []) memMap[s.container] = s.points ?? []
        setMemSeries(memMap)
        const top = sorted.slice(0, 8).map((w) => w.container)
        setStackedCpu((cpu.series ?? []).filter((s) => top.includes(s.container)))
        setError(null)
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : "Erro")
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

  const gridRows = useMemo(() => rows.slice(0, GRID_LIMIT), [rows])

  return (
    <div className="space-y-6">
      <PageHeader
        title="Workloads"
        description="Containers monitorados pelo agent — visão gráfica e tabela."
        breadcrumb="Infraestrutura"
        actions={
          <div className="flex gap-1">
            <Button size="sm" variant={view === "grid" ? "default" : "outline"} onClick={() => setView("grid")}>
              Grid
            </Button>
            <Button size="sm" variant={view === "table" ? "default" : "outline"} onClick={() => setView("table")}>
              Tabela
            </Button>
          </div>
        }
      />

      {error ? <p className="text-destructive text-sm">{error}</p> : null}

      <StackedAreaChart
        title="Top containers — CPU empilhado"
        description="Oito workloads com maior uso no último intervalo"
        series={stackedCpu}
        loading={loading}
        unit="%"
      />

      {view === "grid" ? (
        <div className="space-y-3">
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
            {loading
              ? Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-48 w-full" />)
              : gridRows.map((w) => (
                  <ContainerMetricCard
                    key={w.entity_uid}
                    workload={w}
                    cpuPoints={cpuSeries[w.container]}
                    memPoints={memSeries[w.container]}
                  />
                ))}
          </div>
          {rows.length > GRID_LIMIT ? (
            <p className="text-muted-foreground text-sm">
              Mostrando {GRID_LIMIT} de {rows.length} containers.{" "}
              <button type="button" className="text-primary hover:underline" onClick={() => setView("table")}>
                Ver todos na tabela
              </button>
            </p>
          ) : null}
        </div>
      ) : (
        <ScrollArea className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Container</TableHead>
                <TableHead>CPU</TableHead>
                <TableHead>Memória</TableHead>
                <TableHead>Atualizado</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={5}>
                    <Skeleton className="h-8 w-full" />
                  </TableCell>
                </TableRow>
              ) : rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-muted-foreground text-center">
                    Nenhum container monitorado. Suba o agent com acesso ao Docker socket.
                  </TableCell>
                </TableRow>
              ) : (
                rows.map((w) => (
                  <TableRow key={w.entity_uid}>
                    <TableCell className="font-medium">{w.container}</TableCell>
                    <TableCell>
                      <Badge variant={w.cpu_usage > 80 ? "destructive" : "secondary"}>
                        {formatPercent(w.cpu_usage)}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-sm">
                      {formatBytes(w.memory_usage)}
                      {w.memory_limit > 0 ? (
                        <span className="text-muted-foreground"> / {formatBytes(w.memory_limit)}</span>
                      ) : null}
                    </TableCell>
                    <TableCell className="text-muted-foreground text-xs whitespace-nowrap">
                      {new Date(w.updated_at).toLocaleString("pt-BR")}
                    </TableCell>
                    <TableCell>
                      <Link
                        to={`/metrics?container=${encodeURIComponent(w.container)}`}
                        className="text-primary text-sm hover:underline"
                      >
                        Gráficos
                      </Link>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </ScrollArea>
      )}
    </div>
  )
}
