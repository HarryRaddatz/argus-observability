import { useCallback, useEffect, useMemo, useState } from "react"
import { useSearchParams } from "react-router-dom"

import { MultiSeriesChart } from "@/components/metrics/multi-series-chart"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { cn } from "@/lib/utils"
import {
  fetchMetricCatalog,
  fetchMetricSeries,
  listWorkloads,
  listWorkloadGroups,
  type ContainerSeries,
  type MetricCatalogEntry,
} from "@/lib/api"
import {
  deleteView,
  loadSavedViews,
  saveView,
  TIME_RANGES,
  type SavedView,
} from "@/lib/observability"

export function ExplorerPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const groupId = searchParams.get("group") ?? ""

  const [catalog, setCatalog] = useState<MetricCatalogEntry[]>([])
  const [containers, setContainers] = useState<string[]>([])
  const [groups, setGroups] = useState<{ id: string; name: string }[]>([])
  const [selectedContainers, setSelectedContainers] = useState<string[]>([])
  const [metric, setMetric] = useState(() => searchParams.get("metric") ?? "memory.usage_pct")
  const [since, setSince] = useState("1h")
  const [chartType, setChartType] = useState<"area" | "line">("area")
  const [series, setSeries] = useState<ContainerSeries[]>([])
  const [loading, setLoading] = useState(false)
  const [viewName, setViewName] = useState("")
  const [savedViews, setSavedViews] = useState<SavedView[]>(loadSavedViews)

  useEffect(() => {
    fetchMetricCatalog().then(setCatalog).catch(() => setCatalog([]))
    Promise.all([listWorkloads("1h"), listWorkloadGroups()])
      .then(([wl, g]) => {
        const names = wl.map((w) => w.container).sort()
        setContainers(names)
        setGroups(g.map((x) => ({ id: x.id, name: x.name })))
        if (selectedContainers.length === 0 && names.length > 0 && !groupId) {
          setSelectedContainers(names.slice(0, 3))
        }
      })
      .catch(() => setContainers([]))
  }, [groupId])

  const loadChart = useCallback(() => {
    setLoading(true)
    fetchMetricSeries(metric, since, undefined, groupId || undefined)
      .then((resp) => {
        let s = resp.series ?? []
        if (!groupId && selectedContainers.length > 0) {
          s = s.filter((x) => selectedContainers.includes(x.container))
        }
        setSeries(s)
      })
      .catch(() => setSeries([]))
      .finally(() => setLoading(false))
  }, [metric, since, selectedContainers, groupId])

  useEffect(() => {
    loadChart()
  }, [loadChart])

  const unit = useMemo(() => {
    const m = catalog.find((c) => c.name === metric)
    if (!m) return ""
    if (m.unit === "%") return "%"
    if (m.unit === "bytes") return " MiB"
    return ` ${m.unit}`
  }, [catalog, metric])

  const transform = useMemo(() => {
    if (metric === "memory.usage" || metric === "memory.limit") {
      return (v: number) => Math.round(v / 1024 / 1024)
    }
    if (metric === "http.error_rate") {
      return (v: number) => Math.round(v * 1000) / 10
    }
    if (metric === "http.duration_ms") {
      return (v: number) => Math.round(v)
    }
    return undefined
  }, [metric])

  function toggleContainer(name: string) {
    setSelectedContainers((prev) =>
      prev.includes(name) ? prev.filter((c) => c !== name) : [...prev, name],
    )
  }

  function handleSaveView() {
    if (!viewName.trim()) return
    const view: SavedView = {
      id: crypto.randomUUID(),
      name: viewName.trim(),
      metric,
      containers: selectedContainers,
      since,
      chartType,
      createdAt: new Date().toISOString(),
    }
    saveView(view)
    setSavedViews(loadSavedViews())
    setViewName("")
  }

  function applyView(view: SavedView) {
    setMetric(view.metric)
    setSince(view.since)
    setSelectedContainers(view.containers)
    setChartType(view.chartType)
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Explorer</h1>
        <p className="text-muted-foreground text-sm">
          Monte visualizações customizadas — métricas, containers e período à sua escolha.
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-[280px_1fr]">
        <aside className="space-y-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm">Métrica</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <select
                className="border-input bg-background w-full rounded-md border px-3 py-2 text-sm"
                value={metric}
                onChange={(e) => setMetric(e.target.value)}
              >
                {catalog.map((m) => (
                  <option key={m.name} value={m.name}>
                    {m.label}
                  </option>
                ))}
              </select>
              <div className="flex flex-wrap gap-1">
                {TIME_RANGES.map((r) => (
                  <Button
                    key={r.id}
                    size="sm"
                    variant={since === r.id ? "default" : "outline"}
                    onClick={() => setSince(r.id)}
                  >
                    {r.label}
                  </Button>
                ))}
              </div>
              <div className="flex gap-1">
                <Button
                  size="sm"
                  variant={chartType === "area" ? "default" : "outline"}
                  onClick={() => setChartType("area")}
                >
                  Área
                </Button>
                <Button
                  size="sm"
                  variant={chartType === "line" ? "default" : "outline"}
                  onClick={() => setChartType("line")}
                >
                  Linha
                </Button>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm">Grupo</CardTitle>
            </CardHeader>
            <CardContent>
              <select
                className="border-input bg-background w-full rounded-md border px-3 py-2 text-sm"
                value={groupId}
                onChange={(e) => {
                  const next = e.target.value
                  const p = new URLSearchParams(searchParams)
                  if (next) p.set("group", next)
                  else p.delete("group")
                  setSearchParams(p, { replace: true })
                }}
              >
                <option value="">Nenhum (seleção manual)</option>
                {groups.map((g) => (
                  <option key={g.id} value={g.id}>
                    {g.name}
                  </option>
                ))}
              </select>
              {groupId ? (
                <p className="text-muted-foreground mt-2 text-xs">
                  Filtrando pelo grupo — containers abaixo desabilitados.
                </p>
              ) : null}
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm">Containers</CardTitle>
            </CardHeader>
            <CardContent>
              <ScrollArea className="h-48">
                <ul className="space-y-1">
                  {containers.map((name) => (
                    <li key={name}>
                      <button
                        type="button"
                        disabled={Boolean(groupId)}
                        className={cn(
                          "hover:bg-muted w-full rounded px-2 py-1.5 text-left text-xs disabled:opacity-40",
                          selectedContainers.includes(name) && "bg-muted font-medium",
                        )}
                        onClick={() => toggleContainer(name)}
                      >
                        {name}
                      </button>
                    </li>
                  ))}
                </ul>
              </ScrollArea>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm">Salvar visualização</CardTitle>
            </CardHeader>
            <CardContent className="flex gap-2">
              <Input
                placeholder="Nome..."
                value={viewName}
                onChange={(e) => setViewName(e.target.value)}
              />
              <Button size="sm" onClick={handleSaveView}>
                Salvar
              </Button>
            </CardContent>
          </Card>

          {savedViews.length > 0 ? (
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm">Salvas</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                {savedViews.map((v) => (
                  <div key={v.id} className="flex items-center gap-1">
                    <Button
                      size="sm"
                      variant="ghost"
                      className="h-auto flex-1 justify-start py-1 text-xs"
                      onClick={() => applyView(v)}
                    >
                      {v.name}
                    </Button>
                    <Button size="sm" variant="ghost" onClick={() => {
                      deleteView(v.id)
                      setSavedViews(loadSavedViews())
                    }}>
                      x
                    </Button>
                  </div>
                ))}
              </CardContent>
            </Card>
          ) : null}
        </aside>

        <MultiSeriesChart
          title={catalog.find((c) => c.name === metric)?.label ?? metric}
          description={`${selectedContainers.length || "Todos"} container(s) — ${since}`}
          series={series}
          loading={loading}
          unit={unit}
          transform={transform}
          chartType={chartType}
        />
      </div>
    </div>
  )
}
