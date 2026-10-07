import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"

import { GroupSelect } from "@/components/filters/group-select"
import { TimeRangePicker } from "@/components/filters/time-range-picker"
import { ListPager } from "@/components/list-pager"
import { PageHeader } from "@/components/layout/page-header"
import { ContainerMetricCard } from "@/components/metrics/container-metric-card"
import { StatCard } from "@/components/metrics/stat-card"
import { StackedAreaChart } from "@/components/metrics/stacked-area-chart"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { usePolling } from "@/hooks/use-polling"
import { useQueryState } from "@/hooks/use-query-state"
import {
  fetchFleetStatus,
  fetchMetricSeries,
  getWorkloadGroupSummary,
  listWorkloadGroups,
  listWorkloads,
  type ContainerFleetStatus,
  type ContainerSeries,
  type FleetStatusResponse,
  type SeriesPoint,
  type WorkloadGroup,
  type WorkloadSnapshot,
} from "@/lib/api"
import { instabilityReason, isUnstable, stateLabel, stateVariant } from "@/lib/container-state"
import { formatBytes, formatPercent } from "@/lib/format"
import { PAGE_SIZE, pageIndex, pageSlice } from "@/lib/page"
import { GroupsPanel } from "@/views/groups-panel"

const GRID_LIMIT = 40

type Row = {
  container: string
  workload?: WorkloadSnapshot
  status?: ContainerFleetStatus
}

export function ContainersPage() {
  const { t } = useTranslation()
  const [view, setView] = useQueryState("view", "grid")
  const [group, setGroup] = useQueryState("group", "")
  const [panel, setPanel] = useQueryState("panel", "")
  const [filter, setFilter] = useQueryState("filter", "")
  const [pageRaw, setPage] = useQueryState("page", "1")
  const [since, setSince] = useQueryState("since", "1h")
  const filterKey = `${group}|${filter}`
  const filterSeen = useRef(filterKey)
  useEffect(() => {
    if (filterSeen.current === filterKey) return
    filterSeen.current = filterKey
    setPage("1")
  }, [filterKey, setPage])

  const [workloads, setWorkloads] = useState<WorkloadSnapshot[]>([])
  const [fleet, setFleet] = useState<FleetStatusResponse | null>(null)
  const [cpuSeries, setCpuSeries] = useState<ContainerSeries[]>([])
  const [memSeries, setMemSeries] = useState<ContainerSeries[]>([])
  const [groups, setGroups] = useState<WorkloadGroup[]>([])
  const [members, setMembers] = useState<Set<string> | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    Promise.all([
      listWorkloads("30m"),
      fetchFleetStatus(),
      fetchMetricSeries("cpu.usage", since),
      fetchMetricSeries("memory.usage_pct", since),
    ])
      .then(([wl, fl, cpu, mem]) => {
        setWorkloads(wl)
        setFleet(fl)
        setCpuSeries(cpu.series ?? [])
        setMemSeries(mem.series ?? [])
        setError(null)
      })
      .catch((e) => setError(e instanceof Error ? e.message : t("containers.loadError")))
      .finally(() => setLoading(false))
  }, [since, t])

  usePolling(load)

  const loadGroups = useCallback(() => {
    listWorkloadGroups().then(setGroups).catch(() => setGroups([]))
  }, [])

  useEffect(() => {
    loadGroups()
  }, [loadGroups])

  useEffect(() => {
    if (!group) {
      setMembers(null)
      return
    }
    getWorkloadGroupSummary(group, "30m")
      .then((s) => setMembers(new Set(s.members.map((m) => m.container))))
      .catch(() => setMembers(new Set()))
  }, [group])

  const rows = useMemo(() => {
    const byName = new Map<string, Row>()
    for (const w of workloads) byName.set(w.container, { container: w.container, workload: w })
    for (const s of fleet?.containers ?? []) {
      const row = byName.get(s.container) ?? { container: s.container }
      row.status = s
      byName.set(s.container, row)
    }
    let list = [...byName.values()]
    if (members) list = list.filter((r) => members.has(r.container))
    if (filter === "unstable") list = list.filter((r) => r.status && isUnstable(r.status))
    return list.sort((a, b) => {
      const ua = a.status && isUnstable(a.status) ? 1 : 0
      const ub = b.status && isUnstable(b.status) ? 1 : 0
      if (ua !== ub) return ub - ua
      return (b.workload?.cpu_usage ?? 0) - (a.workload?.cpu_usage ?? 0)
    })
  }, [workloads, fleet, members, filter])

  const counts = useMemo(() => {
    let running = 0
    let unstable = 0
    let stopped = 0
    let restarts = 0
    for (const r of rows) {
      const state = r.status?.state?.toLowerCase()
      if (state === "running") running++
      if (state === "exited" || state === "dead") stopped++
      if (r.status && isUnstable(r.status)) unstable++
      restarts += r.status?.restart_count ?? 0
    }
    return { running, unstable, stopped, restarts }
  }, [rows])

  const seriesByName = useMemo(() => {
    const toMap = (series: ContainerSeries[]) => {
      const m: Record<string, SeriesPoint[]> = {}
      for (const s of series) m[s.container] = s.points ?? []
      return m
    }
    return { cpu: toMap(cpuSeries), mem: toMap(memSeries) }
  }, [cpuSeries, memSeries])

  const stackedCpu = useMemo(() => {
    const top = rows
      .filter((r) => r.workload)
      .sort((a, b) => (b.workload?.cpu_usage ?? 0) - (a.workload?.cpu_usage ?? 0))
      .slice(0, 8)
      .map((r) => r.container)
    return cpuSeries.filter((s) => top.includes(s.container))
  }, [rows, cpuSeries])

  const tablePage = useMemo(() => pageSlice(rows, pageIndex(pageRaw), PAGE_SIZE), [rows, pageRaw])
  const gridRows = rows.filter((r) => r.workload).slice(0, GRID_LIMIT)
  const allContainers = useMemo(() => workloads.map((w) => w.container).sort(), [workloads])

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("containers.title")}
        actions={
          <>
            <GroupSelect value={group} onChange={setGroup} groups={groups} />
            <TimeRangePicker value={since} onChange={setSince} />
            <Button size="sm" variant="outline" onClick={() => setPanel("groups")}>
              {t("containers.manageGroups")}
            </Button>
            <Tabs value={view} onValueChange={(v) => setView(String(v))}>
              <TabsList aria-label={t("containers.view")}>
                <TabsTrigger value="grid">{t("containers.grid")}</TabsTrigger>
                <TabsTrigger value="table">{t("containers.table")}</TabsTrigger>
              </TabsList>
            </Tabs>
          </>
        }
      />

      {error ? <p className="text-destructive text-sm">{error}</p> : null}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label={t("containers.running")} value={counts.running} loading={loading} />
        <StatCard
          label={t("containers.withIssues")}
          value={counts.unstable}
          tone={counts.unstable > 0 ? "critical" : "default"}
          hint={counts.unstable > 0 && filter !== "unstable" ? t("containers.seeThose") : undefined}
          to={
            counts.unstable > 0 && filter !== "unstable"
              ? `/containers?${new URLSearchParams({ ...(group ? { group } : {}), filter: "unstable", view: "table" })}`
              : undefined
          }
          loading={loading}
        />
        <StatCard label={t("containers.stopped")} value={counts.stopped} loading={loading} />
        <StatCard
          label={t("containers.restarts")}
          value={counts.restarts}
          tone={counts.restarts > 10 ? "warning" : "default"}
          loading={loading}
        />
      </div>

      {filter === "unstable" ? (
        <div className="bg-muted/50 flex flex-wrap items-center gap-2 rounded-md border px-3 py-2 text-sm">
          <span>{t("containers.unstableBanner")}</span>
          <Button size="sm" variant="ghost" onClick={() => setFilter("")}>
            {t("containers.showAll")}
          </Button>
        </div>
      ) : null}

      {view === "grid" ? (
        <div className="space-y-4">
          <StackedAreaChart
            title={t("containers.cpuTopTitle")}
            description={t("containers.cpuTopOf", {
              shown: stackedCpu.length,
              total: rows.filter((r) => r.workload).length,
            })}
            series={stackedCpu}
            loading={loading}
            unit="%"
          />
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
            {loading
              ? Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-48 w-full" />)
              : gridRows.map((r) => (
                  <ContainerMetricCard
                    key={r.container}
                    workload={r.workload!}
                    status={r.status}
                    cpuPoints={seriesByName.cpu[r.container]}
                    memPoints={seriesByName.mem[r.container]}
                  />
                ))}
          </div>
          {!loading && gridRows.length === 0 ? <EmptyRows filtered={Boolean(group || filter)} /> : null}
          {rows.length > gridRows.length ? (
            <p className="text-muted-foreground text-sm">
              {t("containers.gridShows", { shown: gridRows.length, total: rows.length })}{" "}
              <button type="button" className="text-primary hover:underline" onClick={() => setView("table")}>
                {t("containers.seeAllTable")}
              </button>
            </p>
          ) : null}
        </div>
      ) : (
        <div className="space-y-6">
          <ScrollArea className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("common.container")}</TableHead>
                  <TableHead>{t("containers.state")}</TableHead>
                  <TableHead>{t("containers.cpu")}</TableHead>
                  <TableHead>{t("containers.memory")}</TableHead>
                  <TableHead>{t("containers.restarts")}</TableHead>
                  <TableHead>{t("containers.attention")}</TableHead>
                  <TableHead>
                    <span className="sr-only">{t("containers.actions")}</span>
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading && rows.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7}>
                      <Skeleton className="h-8 w-full" />
                    </TableCell>
                  </TableRow>
                ) : rows.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7}>
                      <EmptyRows filtered={Boolean(group || filter)} />
                    </TableCell>
                  </TableRow>
                ) : (
                  tablePage.rows.map((r) => <ContainerRow key={r.container} row={r} />)
                )}
              </TableBody>
            </Table>
          </ScrollArea>
          <ListPager
            start={tablePage.start}
            end={tablePage.end}
            total={tablePage.total}
            page={tablePage.page}
            pageSize={PAGE_SIZE}
            onPage={(n) => setPage(String(n))}
          />

          {fleet && fleet.services.length > 0 && !group ? (
            <section className="space-y-2">
              <h2 className="text-sm font-medium">{t("containers.replicasByService")}</h2>
              <ScrollArea className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t("common.service")}</TableHead>
                      <TableHead>{t("containers.replicasUp")}</TableHead>
                      <TableHead>{t("containers.restarting")}</TableHead>
                      <TableHead>{t("containers.healthFailing")}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {fleet.services.map((s) => (
                      <TableRow key={s.service}>
                        <TableCell className="font-medium">{s.service}</TableCell>
                        <TableCell className="tabular-nums">
                          {s.replicas_up} {t("containers.of")} {s.replicas_total}
                        </TableCell>
                        <TableCell className="tabular-nums">{s.restarting || "—"}</TableCell>
                        <TableCell className="tabular-nums">{s.unhealthy || "—"}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </ScrollArea>
            </section>
          ) : null}
        </div>
      )}

      <GroupsPanel
        open={panel === "groups"}
        onOpenChange={(open) => setPanel(open ? "groups" : "")}
        groups={groups}
        containers={allContainers}
        activeGroup={group}
        onSelectGroup={setGroup}
        onChanged={loadGroups}
      />
    </div>
  )
}

function EmptyRows({ filtered }: { filtered: boolean }) {
  const { t } = useTranslation()
  return (
    <p className="text-muted-foreground py-6 text-center text-sm">
      {filtered ? t("containers.empty") : t("containers.emptyAgent")}
    </p>
  )
}

function ContainerRow({ row }: { row: Row }) {
  const { t } = useTranslation()
  const { workload: w, status } = row
  const state = status?.state?.toLowerCase()
  const unstable = status ? isUnstable(status) : false
  return (
    <TableRow className={unstable ? "bg-destructive/5" : undefined}>
      <TableCell className="font-medium">{row.container}</TableCell>
      <TableCell>
        {status?.disposition ? (
          <Badge variant={status.disposition === "intentional" ? "outline" : "destructive"}>
            {stateLabel(t, status.disposition)}
          </Badge>
        ) : state ? (
          <Badge variant={stateVariant(state)}>{stateLabel(t, state)}</Badge>
        ) : (
          "—"
        )}
      </TableCell>
      <TableCell className="tabular-nums">
        {w ? (
          <Badge variant={w.cpu_usage > 80 ? "destructive" : "secondary"}>{formatPercent(w.cpu_usage)}</Badge>
        ) : (
          "—"
        )}
      </TableCell>
      <TableCell className="text-sm tabular-nums">
        {w ? (
          <>
            {formatBytes(w.memory_usage)}
            {w.memory_limit > 0 ? <span className="text-muted-foreground"> {t("containers.of")} {formatBytes(w.memory_limit)}</span> : null}
          </>
        ) : (
          "—"
        )}
      </TableCell>
      <TableCell className={status && status.restart_count > 3 ? "font-medium text-amber-600 tabular-nums" : "tabular-nums"}>
        {status?.restart_count ?? "—"}
      </TableCell>
      <TableCell className="text-sm">{unstable && status ? instabilityReason(t, status) : "—"}</TableCell>
      <TableCell>
        <div className="flex gap-3 text-xs">
          <Link to={`/metrics?container=${encodeURIComponent(row.container)}`} className="text-primary hover:underline">
            {t("containers.metricsLink")}
          </Link>
          <Link to={`/logs?container=${encodeURIComponent(row.container)}`} className="text-primary hover:underline">
            {t("containers.logsLink")}
          </Link>
        </div>
      </TableCell>
    </TableRow>
  )
}
