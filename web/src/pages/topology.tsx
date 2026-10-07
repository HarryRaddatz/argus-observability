import { useCallback, useEffect, useId, useMemo, useState } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"

import { TimeRangePicker } from "@/components/filters/time-range-picker"
import { PageHeader } from "@/components/layout/page-header"
import { Badge } from "@/components/ui/badge"
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
  type EdgeView,
  type OpStatus,
  type ServiceView,
} from "@/lib/topology-view"

const emptyGraph: TopologyGraph = { nodes: [], edges: [] }

const statusClass: Record<OpStatus, string> = {
  critical: "border-destructive",
  warning: "border-foreground",
  healthy: "border-border",
  unknown: "border-border",
}

export function TopologyPage() {
  const { t } = useTranslation()
  const searchId = useId()
  const [since, setSince] = useQueryState("since", "24h")
  const [focus, setFocus] = useQueryState("focus", "")
  const [query, setQuery] = useQueryState("q", "")
  const [graph, setGraph] = useState<TopologyGraph>(emptyGraph)
  const [fleet, setFleet] = useState<ContainerFleetStatus[]>([])
  const [workloads, setWorkloads] = useState<WorkloadSnapshot[]>([])
  const [http, setHttp] = useState<HTTPServiceSummary[]>([])
  const [alerts, setAlerts] = useState<ActiveAlert[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const load = useCallback(() => {
    setError(false)
    Promise.all([fetchTopology(since), fetchFleetStatus(), listWorkloads(since), fetchHTTPSummary(since), fetchActiveAlerts()])
      .then(([nextGraph, nextFleet, nextWorkloads, nextHttp, nextAlerts]) => {
        setGraph(nextGraph)
        setFleet(nextFleet.containers ?? [])
        setWorkloads(nextWorkloads)
        setHttp(nextHttp)
        setAlerts(nextAlerts)
      })
      .catch(() => setError(true))
      .finally(() => setLoading(false))
  }, [since])

  useEffect(() => {
    load()
  }, [load])

  const ranked = useMemo(
    () => rankServices(buildServices(graph, { fleet, workloads, http, alerts })),
    [graph, fleet, workloads, http, alerts],
  )
  const visible = useMemo(() => filterServices(ranked, query), [ranked, query])
  const selected = pickService(focus ? ranked.filter((service) => service.id === focus) : ranked, focus)
  const counts = useMemo(() => {
    const tally = { critical: 0, warning: 0, healthy: 0, unknown: 0 }
    for (const service of ranked) tally[service.status] += 1
    return tally
  }, [ranked])

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("topology.title")}
        description={t("topology.description")}
        actions={<TimeRangePicker value={since} onChange={setSince} />}
      />

      {loading && ranked.length === 0 ? (
        <Skeleton className="h-48 w-full" />
      ) : error && ranked.length === 0 ? (
        <p className="text-destructive text-sm">{t("topology.loadError")}</p>
      ) : ranked.length === 0 ? (
        <p className="text-muted-foreground text-sm">{t("topology.empty")}</p>
      ) : (
        <>
          <p className="text-sm" aria-live="polite">
            {t("topology.summary", { critical: counts.critical, warning: counts.warning, total: ranked.length })}
          </p>
          <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
            <div className="space-y-6">
              <div className="space-y-2">
                <label htmlFor={searchId} className="text-sm font-medium">
                  {t("topology.search")}
                </label>
                <Input id={searchId} value={query} onChange={(event) => setQuery(event.target.value)} />
                <p className="text-muted-foreground text-xs tabular-nums">
                  {t("topology.shown", { shown: visible.length, total: ranked.length })}
                </p>
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
                        className="hover:bg-muted/50 focus-visible:ring-ring aria-pressed:bg-muted flex min-h-11 w-full items-center justify-between gap-3 px-3 py-2 text-left focus-visible:ring-2 focus-visible:outline-none"
                      >
                        <span className="font-medium">{service.label}</span>
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
        </>
      )}
    </div>
  )
}

function StatusBadge({ status }: { status: OpStatus }) {
  const { t } = useTranslation()
  const variant = status === "critical" ? "destructive" : status === "warning" ? "secondary" : "outline"
  return <Badge variant={variant}>{t(`topology.status.${status}`)}</Badge>
}

function Path({ service, onSelect }: { service: ServiceView; onSelect: (id: string) => void }) {
  const { t } = useTranslation()
  return (
    <section className="space-y-2" aria-label={t("topology.path")}>
      <h2 className="text-sm font-medium">{t("topology.path")}</h2>
      <div className="flex flex-wrap items-center gap-2">
        <EdgeGroup title={t("topology.upstream")} edges={service.upstream} onSelect={onSelect} />
        <span aria-hidden="true" className="text-muted-foreground">
          →
        </span>
        <span className={`rounded-md border px-3 py-2 text-sm font-medium ${statusClass[service.status]}`}>{service.label}</span>
        <span aria-hidden="true" className="text-muted-foreground">
          →
        </span>
        <EdgeGroup title={t("topology.downstream")} edges={service.downstream} onSelect={onSelect} />
      </div>
    </section>
  )
}

function EdgeGroup({ title, edges, onSelect }: { title: string; edges: EdgeView[]; onSelect: (id: string) => void }) {
  const { t } = useTranslation()
    if (edges.length === 0) {
    return <p className="text-muted-foreground text-sm">{title === t("topology.upstream") ? t("topology.noUpstream") : t("topology.noDownstream")}</p>
  }
  return (
    <ul className="flex flex-wrap gap-2" aria-label={title}>
      {edges.map((edge) => (
        <li key={`${edge.other}-${edge.kind}-${edge.port ?? ""}`}>
          <button type="button" onClick={() => onSelect(edge.other)} className="hover:bg-muted rounded-md border px-2 py-1 text-left text-sm">
            <span className="block font-medium">{edge.other}</span>
            <span className="text-muted-foreground block text-xs">
              {t("topology.edgeMeta", {
                kind: edge.kind,
                count: edge.count,
                origin: edge.origin || t("topology.originUnknown"),
              })}
            </span>
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
    <aside className="space-y-4 rounded-md border p-4" aria-label={t("topology.detail")}>
      <div className="space-y-1">
        <h2 className="text-base font-medium">{service.label}</h2>
        <StatusBadge status={service.status} />
      </div>
      <dl className="space-y-2 text-sm">
        {workload ? (
          <>
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
            {service.workloads.length > 1 ? (
              <p className="text-muted-foreground text-xs">{t("topology.cpuOf", { name: workload.container, count: service.workloads.length })}</p>
            ) : null}
          </>
        ) : (
          <p className="text-muted-foreground text-xs">{t("topology.noWorkload")}</p>
        )}
        {service.http ? (
          <>
            <div className="flex justify-between gap-3">
              <dt className="text-muted-foreground">{t("topology.errorRate")}</dt>
              <dd className="tabular-nums">{formatPercent(service.http.error_rate * 100)}</dd>
            </div>
            <div className="flex justify-between gap-3">
              <dt className="text-muted-foreground">{t("topology.latency")}</dt>
              <dd className="tabular-nums">{t("topology.latencyMs", { ms: Math.round(service.http.avg_latency_ms) })}</dd>
            </div>
          </>
        ) : (
          <p className="text-muted-foreground text-xs">{t("topology.noHttp")}</p>
        )}
        {service.fleet.length === 0 ? <p className="text-muted-foreground text-xs">{t("topology.noFleet")}</p> : null}
      </dl>
      <div className="flex flex-col gap-2 text-sm">
        <Link className="text-primary hover:underline" to={`/logs?container=${encodeURIComponent(service.id)}&since=${since}`}>
          {t("topology.errorsIn", { name: service.label })}
        </Link>
        <Link className="text-primary hover:underline" to={`/traces?service=${encodeURIComponent(service.id)}&since=${since}`}>
          {t("topology.tracesOf", { name: service.label })}
        </Link>
        <Link className="text-primary hover:underline" to={`/metrics?container=${encodeURIComponent(service.id)}&since=${since}`}>
          {t("topology.metricsOf", { name: service.label })}
        </Link>
      </div>
    </aside>
  )
}
