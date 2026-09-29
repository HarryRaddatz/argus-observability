import { useEffect, useMemo, useRef, useState } from "react"
import { Link } from "react-router-dom"

import { MultiSeriesChart } from "@/components/metrics/multi-series-chart"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useQueryPatch, useQueryState } from "@/hooks/use-query-state"
import { fetchMetricCatalog, fetchMetricSeries, type ContainerSeries, type MetricCatalogEntry, type WorkloadSnapshot } from "@/lib/api"
import { deleteView, loadSavedViews, matchesService, saveView, type SavedView } from "@/lib/observability"
import { cn } from "@/lib/utils"

type Props = {
  since: string
  group: string
  workloads: WorkloadSnapshot[]
}

function seriesContainer(s: ContainerSeries) {
  return s.entity_uid.split(":").pop() || s.container
}

export function MetricCompare({ since, group, workloads }: Props) {
  const [metric, setMetric] = useQueryState("metric", "memory.usage_pct")
  const [chartType, setChartType] = useQueryState("chart", "area")
  const [stat, setStat] = useQueryState("stat", "avg")
  const [containersParam] = useQueryState("containers", "")
  const [service] = useQueryState("service", "")
  const patch = useQueryPatch()

  const [catalog, setCatalog] = useState<MetricCatalogEntry[]>([])
  const [series, setSeries] = useState<ContainerSeries[]>([])
  const [loading, setLoading] = useState(false)
  const [viewName, setViewName] = useState("")
  const [savedViews, setSavedViews] = useState<SavedView[]>(loadSavedViews)

  const initialized = useRef(false)
  const selected = useMemo(() => containersParam.split(",").filter(Boolean), [containersParam])
  const names = useMemo(() => workloads.map((w) => w.container).sort(), [workloads])

  useEffect(() => {
    fetchMetricCatalog().then(setCatalog).catch(() => setCatalog([]))
  }, [])

  useEffect(() => {
    if (initialized.current || names.length === 0) return
    initialized.current = true
    if (group || selected.length > 0) return
    const initial = service
      ? workloads.filter((w) => matchesService(w.container, w.service, service)).map((w) => w.container)
      : names.slice(0, 3)
    patch({ containers: (initial.length > 0 ? initial : names.slice(0, 3)).join(","), service: null })
  }, [group, selected.length, names, service, workloads, patch])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    fetchMetricSeries(metric, since, undefined, group || undefined)
      .then((resp) => {
        if (cancelled) return
        let s = resp.series ?? []
        if (!group && selected.length > 0) {
          s = s.filter((x) => selected.includes(x.container) || selected.includes(seriesContainer(x)))
        }
        setSeries(s)
      })
      .catch(() => {
        if (!cancelled) setSeries([])
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [metric, since, group, selected])

  const unit = useMemo(() => {
    const m = catalog.find((c) => c.name === metric)
    if (!m) return ""
    if (m.unit === "%") return "%"
    if (m.unit === "bytes") return " MiB"
    return ` ${m.unit}`
  }, [catalog, metric])

  const transform = useMemo(() => {
    if (metric === "memory.usage" || metric === "memory.limit") return (v: number) => Math.round(v / 1024 / 1024)
    if (metric === "http.error_rate") return (v: number) => Math.round(v * 1000) / 10
    if (metric === "http.duration_ms") return (v: number) => Math.round(v)
    return undefined
  }, [metric])

  function toggleContainer(name: string) {
    const next = selected.includes(name) ? selected.filter((c) => c !== name) : [...selected, name]
    patch({ containers: next.join(",") })
  }

  function handleSaveView() {
    if (!viewName.trim()) return
    saveView({
      id: crypto.randomUUID(),
      name: viewName.trim(),
      metric,
      containers: selected,
      since,
      chartType: chartType === "line" ? "line" : "area",
      createdAt: new Date().toISOString(),
    })
    setSavedViews(loadSavedViews())
    setViewName("")
  }

  function applyView(view: SavedView) {
    patch({ metric: view.metric, since: view.since, containers: view.containers.join(","), chart: view.chartType, group: null })
  }

  const openable = useMemo(
    () => (group || selected.length === 0 ? [...new Set(series.map(seriesContainer))] : selected),
    [group, selected, series],
  )
  const metricLabel = catalog.find((c) => c.name === metric)?.label ?? metric
  const scope = group
    ? "Containers do grupo"
    : selected.length === 0
      ? "Todos os containers"
      : selected.length === 1
        ? "1 container"
        : `${selected.length} containers`

  return (
    <div className="grid gap-6 lg:grid-cols-[280px_1fr]">
      <aside className="space-y-5">
        <section className="space-y-2">
          <h2 className="text-sm font-medium">Métrica</h2>
          <select
            aria-label="Métrica"
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
          <div className="flex flex-wrap gap-2">
            <div role="group" aria-label="Resumo" className="flex gap-1">
              <Button size="sm" variant={stat === "avg" ? "default" : "outline"} aria-pressed={stat === "avg"} onClick={() => setStat("avg")}>
                Média
              </Button>
              <Button size="sm" variant={stat === "max" ? "default" : "outline"} aria-pressed={stat === "max"} onClick={() => setStat("max")}>
                Pico
              </Button>
            </div>
            <div role="group" aria-label="Tipo de gráfico" className="flex gap-1">
              <Button size="sm" variant={chartType === "area" ? "default" : "outline"} aria-pressed={chartType === "area"} onClick={() => setChartType("area")}>
                Área
              </Button>
              <Button size="sm" variant={chartType === "line" ? "default" : "outline"} aria-pressed={chartType === "line"} onClick={() => setChartType("line")}>
                Linha
              </Button>
            </div>
          </div>
        </section>

        <section className="space-y-2">
          <h2 className="text-sm font-medium">Containers</h2>
          {group ? (
            <p className="text-muted-foreground text-xs">Com um grupo selecionado, o gráfico mostra todos os containers dele.</p>
          ) : null}
          <ScrollArea className="h-48 rounded-md border">
            <ul className="space-y-1 p-1">
              {names.map((name) => (
                <li key={name}>
                  <button
                    type="button"
                    disabled={Boolean(group)}
                    aria-pressed={selected.includes(name)}
                    className={cn(
                      "hover:bg-muted w-full rounded px-2 py-1.5 text-left text-xs disabled:opacity-40",
                      selected.includes(name) && "bg-muted font-medium",
                    )}
                    onClick={() => toggleContainer(name)}
                  >
                    {name}
                  </button>
                </li>
              ))}
            </ul>
          </ScrollArea>
        </section>

        <section className="space-y-2">
          <h2 className="text-sm font-medium">Visões salvas</h2>
          <form
            className="flex gap-2"
            onSubmit={(e) => {
              e.preventDefault()
              handleSaveView()
            }}
          >
            <Input aria-label="Nome da visão" placeholder="Nome da visão" value={viewName} onChange={(e) => setViewName(e.target.value)} />
            <Button size="sm" type="submit" disabled={!viewName.trim()}>
              Salvar
            </Button>
          </form>
          {savedViews.length > 0 ? (
            <ul className="space-y-1">
              {savedViews.map((v) => (
                <li key={v.id} className="flex items-center gap-1">
                  <Button size="sm" variant="ghost" className="h-auto flex-1 justify-start py-1 text-xs" onClick={() => applyView(v)}>
                    {v.name}
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    aria-label={`Apagar a visão ${v.name}`}
                    onClick={() => {
                      deleteView(v.id)
                      setSavedViews(loadSavedViews())
                    }}
                  >
                    Apagar
                  </Button>
                </li>
              ))}
            </ul>
          ) : null}
        </section>
      </aside>

      <div className="min-w-0 space-y-3">
        <MultiSeriesChart
          title={metricLabel}
          description={scope}
          series={series}
          loading={loading}
          unit={unit}
          transform={transform}
          chartType={chartType === "line" ? "line" : "area"}
          statMode={stat === "max" ? "max" : "avg"}
        />
        {openable.length > 0 ? (
          <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm">
            <span className="text-muted-foreground">Abrir um container:</span>
            {openable.map((name) => (
              <Link
                key={name}
                to={`/metrics?container=${encodeURIComponent(name)}&since=${since}`}
                className="text-primary hover:underline"
              >
                {name}
              </Link>
            ))}
          </div>
        ) : null}
      </div>
    </div>
  )
}
