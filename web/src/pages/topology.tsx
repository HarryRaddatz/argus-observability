import { useCallback, useEffect, useId, useMemo, useRef, useState } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"

import { TimeRangePicker } from "@/components/filters/time-range-picker"
import { PageHeader } from "@/components/layout/page-header"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { useQueryState } from "@/hooks/use-query-state"
import {
  fetchActiveAlerts,
  fetchFleetStatus,
  fetchHTTPSummary,
  fetchTopology,
  listWorkloads,
  type ActiveAlert,
  type ContainerFleetStatus,
  type HTTPServiceSummary,
  type TopologyGraph,
  type WorkloadSnapshot,
} from "@/lib/api"
import { formatBytes, formatPercent } from "@/lib/format"
import {
  buildServices,
  busiestWorkload,
  filterServices,
  pickService,
  rankServices,
  splitAddresses,
  type EdgeView,
  type OpStatus,
  type ServiceView,
} from "@/lib/topology-view"

const emptyGraph: TopologyGraph = { nodes: [], edges: [] }

export function TopologyPage() {
  const { t } = useTranslation()
  const searchId = useId()
  const [since, setSince] = useQueryState("since", "24h")
  const [focus, setFocus] = useQueryState("focus", "")
  const [query, setQuery] = useQueryState("q", "")
  const [showAddresses, setShowAddresses] = useQueryState("addresses", "")
  const [graph, setGraph] = useState<TopologyGraph>(emptyGraph)
  const [fleet, setFleet] = useState<ContainerFleetStatus[]>([])
  const [workloads, setWorkloads] = useState<WorkloadSnapshot[]>([])
  const [http, setHttp] = useState<HTTPServiceSummary[]>([])
  const [alerts, setAlerts] = useState<ActiveAlert[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const generation = useRef(0)
  const load = useCallback(() => {
    const run = ++generation.current
    const current = () => run === generation.current
    setLoading(true)
    setError(false)
    fetchTopology(since)
      .then((next) => {
        if (current()) setGraph(next)
      })
      .catch(() => {
        if (current()) setError(true)
      })
      .finally(() => {
        if (current()) setLoading(false)
      })
    const side = <T,>(promise: Promise<T>, apply: (value: T | undefined) => void) =>
      promise.then(
        (value) => current() && apply(value),
        () => current() && apply(undefined),
      )
    side(fetchFleetStatus(), (v) => setFleet(v?.containers ?? []))
    side(listWorkloads(since), (v) => setWorkloads(v ?? []))
    side(fetchHTTPSummary(since), (v) => setHttp(v ?? []))
    side(fetchActiveAlerts(), (v) => setAlerts(v ?? []))
  }, [since])

  useEffect(() => {
    load()
  }, [load])

  const ranked = useMemo(
    () => rankServices(buildServices(graph, { fleet, workloads, http, alerts })),
    [graph, fleet, workloads, http, alerts],
  )
  const { named, addresses } = useMemo(() => splitAddresses(ranked), [ranked])
  const listed = showAddresses ? ranked : named
  const visible = useMemo(() => filterServices(listed, query), [listed, query])
  const focusMissing = focus !== "" && ranked.length > 0 && !ranked.some((service) => service.id === focus)
  const selected = pickService(focusMissing ? named : ranked, focusMissing ? "" : focus) ?? pickService(ranked, "")
  const counts = useMemo(() => {
    const tally = { critical: 0, warning: 0, healthy: 0, unknown: 0 }
    for (const service of named) tally[service.status] += 1
    return tally
  }, [named])

  const retry = (
    <Button variant="outline" size="sm" onClick={load}>
      {t("common.retry")}
    </Button>
  )

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("topology.title")}
        description={t("topology.description")}
        actions={<TimeRangePicker value={since} onChange={setSince} />}
      />

      {loading && ranked.length === 0 ? (
        <Skeleton className="h-48 w-full" aria-label={t("common.loading")} />
      ) : error && ranked.length === 0 ? (
        <div role="alert" className="flex flex-wrap items-center gap-3 text-sm">
          <p className="text-destructive">{t("topology.loadError")}</p>
          {retry}
        </div>
      ) : ranked.length === 0 ? (
        <p className="text-muted-foreground text-sm">{t("topology.empty")}</p>
      ) : (
        <div className="space-y-6" aria-busy={loading}>
          {error ? (
            <div role="alert" className="flex flex-wrap items-center gap-3 text-sm">
              <p className="text-destructive">{t("topology.staleError")}</p>
              {retry}
            </div>
          ) : null}
          <p className="text-sm" aria-live="polite">
            {t("topology.summary", { critical: counts.critical, warning: counts.warning, count: named.length })}
          </p>
          {focusMissing ? (
            <p role="status" className="text-muted-foreground text-sm">
              {t("topology.focusMissing", { name: focus })}
            </p>
          ) : null}
          <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
            <div className="min-w-0 space-y-6">
              <div className="space-y-2">
                <label htmlFor={searchId} className="text-sm font-medium">
                  {t("topology.search")}
                </label>
                <Input
                  id={searchId}
                  type="search"
                  name="service"
                  autoComplete="off"
                  spellCheck={false}
                  placeholder={t("topology.searchPlaceholder")}
                  value={query}
                  onChange={(event) => setQuery(event.target.value)}
                />
                <div className="text-muted-foreground flex flex-wrap items-center justify-between gap-2 text-xs tabular-nums">
                  <span>{t("topology.shown", { shown: visible.length, total: listed.length })}</span>
                  {addresses.length > 0 ? (
                    <Button variant="ghost" size="sm" onClick={() => setShowAddresses(showAddresses ? "" : "1")}>
                      {showAddresses
                        ? t("topology.hideAddresses")
                        : t("topology.showAddresses", { count: addresses.length })}
                    </Button>
                  ) : null}
                </div>
              </div>
              {visible.length === 0 ? (
                <p className="text-muted-foreground text-sm">{t("topology.noMatch")}</p>
              ) : (
                <ul className="divide-y rounded-md border">
                  {visible.map((service) => (
                    <li key={service.id}>
                      <button
                        type="button"
                        aria-pressed={selected?.id === service.id}
                        onClick={() => setFocus(service.id)}
                        className="hover:bg-muted/50 focus-visible:ring-ring aria-pressed:bg-muted flex min-h-11 w-full min-w-0 items-center justify-between gap-3 px-3 py-2 text-left focus-visible:ring-2 focus-visible:outline-none"
                      >
                        <span className="min-w-0 font-medium break-all">{service.label}</span>
                        <StatusBadge status={service.status} />
                      </button>
                    </li>
                  ))}
                </ul>
              )}
              {selected ? <Path service={selected} onSelect={setFocus} /> : null}
            </div>
            {selected ? <Detail service={selected} since={since} /> : null}
          </div>
        </div>
      )}
    </div>
  )
}

function StatusBadge({ status }: { status: OpStatus }) {
  const { t } = useTranslation()
  const variant = status === "critical" ? "destructive" : status === "warning" ? "secondary" : "outline"
  return (
    <Badge variant={variant} className="shrink-0">
      {t(`topology.status.${status}`)}
    </Badge>
  )
}

function Path({ service, onSelect }: { service: ServiceView; onSelect: (id: string) => void }) {
  const { t } = useTranslation()
  return (
    <section className="space-y-2" aria-label={t("topology.path")}>
      <h2 className="text-sm font-medium">{t("topology.path")}</h2>
      <div className="flex flex-wrap items-center gap-2">
        <EdgeGroup edges={service.upstream} empty={t("topology.noUpstream")} label={t("topology.upstream")} onSelect={onSelect} />
        <span aria-hidden="true" className="text-muted-foreground">
          →
        </span>
        <span
          className={`flex min-w-0 items-center gap-2 rounded-md border px-3 py-2 text-sm font-medium ${
            service.status === "critical" ? "border-destructive" : service.status === "warning" ? "border-foreground" : "border-border"
          }`}
        >
          <span className="break-all">{service.label}</span>
          <StatusBadge status={service.status} />
        </span>
        <span aria-hidden="true" className="text-muted-foreground">
          →
        </span>
        <EdgeGroup edges={service.downstream} empty={t("topology.noDownstream")} label={t("topology.downstream")} onSelect={onSelect} />
      </div>
    </section>
  )
}

function EdgeGroup({
  edges,
  empty,
  label,
  onSelect,
}: {
  edges: EdgeView[]
  empty: string
  label: string
  onSelect: (id: string) => void
}) {
  const { t } = useTranslation()
  if (edges.length === 0) {
    return <p className="text-muted-foreground text-sm">{empty}</p>
  }
  return (
    <ul className="flex min-w-0 flex-wrap gap-2" aria-label={label}>
      {edges.map((edge) => (
        <li key={`${edge.other}-${edge.kind}-${edge.port ?? ""}`} className="min-w-0">
          <button
            type="button"
            onClick={() => onSelect(edge.other)}
            className="hover:bg-muted focus-visible:ring-ring min-h-11 rounded-md border px-2 py-1 text-left text-sm focus-visible:ring-2 focus-visible:outline-none"
          >
            <span className="block font-medium break-all">{edge.other}</span>
            <span className="text-muted-foreground block text-xs">
              {t("topology.edgeMeta", {
                kind: t(`topology.kinds.${edge.kind}`, { defaultValue: edge.kind }),
                calls: t("topology.calls", { count: edge.count }),
              })}
            </span>
            <span className="text-muted-foreground block text-xs">{edge.origin || t("topology.originUnknown")}</span>
          </button>
        </li>
      ))}
    </ul>
  )
}

function Detail({ service, since }: { service: ServiceView; since: string }) {
  const { t } = useTranslation()
  const workload = busiestWorkload(service.workloads)
  const memoryPct = workload && workload.memory_limit > 0 ? (workload.memory_usage / workload.memory_limit) * 100 : null
  return (
    <aside className="min-w-0 space-y-4 rounded-md border p-4" aria-label={t("topology.detail")}>
      <div className="space-y-1">
        <h2 className="text-base font-medium break-all">{service.label}</h2>
        <StatusBadge status={service.status} />
      </div>
      <div className="space-y-2 text-sm">
        {workload ? (
          <>
            <dl className="space-y-2">
              <div className="flex justify-between gap-3">
                <dt className="text-muted-foreground">{t("topology.cpu")}</dt>
                <dd className="tabular-nums">{formatPercent(workload.cpu_usage)}</dd>
              </div>
              {memoryPct != null ? (
                <div className="flex justify-between gap-3">
                  <dt className="text-muted-foreground">{t("topology.memory")}</dt>
                  <dd className="tabular-nums">
                    {formatPercent(memoryPct)}
                    <span className="text-muted-foreground"> ({formatBytes(workload.memory_usage)})</span>
                  </dd>
                </div>
              ) : null}
            </dl>
            {service.workloads.length > 1 ? (
              <p className="text-muted-foreground text-xs">{t("topology.cpuOf", { name: workload.container, count: service.workloads.length })}</p>
            ) : null}
          </>
        ) : (
          <p className="text-muted-foreground text-xs">{t("topology.noWorkload")}</p>
        )}
        {service.http ? (
          <dl className="space-y-2">
            <div className="flex justify-between gap-3">
              <dt className="text-muted-foreground">{t("topology.errorRate")}</dt>
              <dd className="tabular-nums">{formatPercent(service.http.error_rate * 100)}</dd>
            </div>
            <div className="flex justify-between gap-3">
              <dt className="text-muted-foreground">{t("topology.latency")}</dt>
              <dd className="tabular-nums">{t("topology.latencyMs", { ms: Math.round(service.http.avg_latency_ms) })}</dd>
            </div>
          </dl>
        ) : (
          <p className="text-muted-foreground text-xs">{t("topology.noHttp")}</p>
        )}
        {service.fleet.length === 0 ? <p className="text-muted-foreground text-xs">{t("topology.noFleet")}</p> : null}
      </div>
      <div className="flex flex-col text-sm">
        <Link className="text-primary inline-flex min-h-11 items-center hover:underline" to={`/logs?container=${encodeURIComponent(service.id)}&since=${since}`}>
          {t("topology.errorsIn", { name: service.label })}
        </Link>
        <Link className="text-primary inline-flex min-h-11 items-center hover:underline" to={`/traces?service=${encodeURIComponent(service.id)}&since=${since}`}>
          {t("topology.tracesOf", { name: service.label })}
        </Link>
        <Link className="text-primary inline-flex min-h-11 items-center hover:underline" to={`/metrics?container=${encodeURIComponent(service.id)}&since=${since}`}>
          {t("topology.metricsOf", { name: service.label })}
        </Link>
      </div>
    </aside>
  )
}
