import { isUnstable } from "@/lib/container-state"
import { matchesService } from "@/lib/observability"
import type {
  ActiveAlert,
  ContainerFleetStatus,
  HTTPServiceSummary,
  TopologyGraph,
  WorkloadSnapshot,
} from "@/lib/api"

export type OpStatus = "critical" | "warning" | "healthy" | "unknown"

export type TopologySignals = {
  fleet: ContainerFleetStatus[]
  workloads: WorkloadSnapshot[]
  http: HTTPServiceSummary[]
  alerts: ActiveAlert[]
}

export type EdgeView = {
  other: string
  kind: string
  count: number
  origin?: string
  port?: number
}

export type ServiceView = {
  id: string
  label: string
  status: OpStatus
  upstream: EdgeView[]
  downstream: EdgeView[]
  fleet: ContainerFleetStatus[]
  workloads: WorkloadSnapshot[]
  http: HTTPServiceSummary | null
  alerts: ActiveAlert[]
}

const rank: Record<OpStatus, number> = { critical: 0, warning: 1, unknown: 2, healthy: 3 }

function same(a: string, b?: string) {
  return !!b && a.toLowerCase() === b.toLowerCase()
}

function matchesNode(nodeId: string, container: string, service?: string) {
  return matchesService(container, service, nodeId)
}

function worse(a: OpStatus, b: OpStatus): OpStatus {
  return rank[a] <= rank[b] ? a : b
}

function fleetStatus(row: ContainerFleetStatus): OpStatus {
  const state = row.state?.toLowerCase() ?? ""
  if (state === "dead" || row.disposition === "oom" || row.disposition === "unexpected" || row.oom_killed) {
    return "critical"
  }
  if (isUnstable(row)) return "warning"
  if (state === "running") return "healthy"
  return "unknown"
}

function alertStatus(alert: ActiveAlert): OpStatus {
  const severity = alert.severity?.toLowerCase() ?? ""
  if (severity === "critical") return "critical"
  if (severity === "warning") return "warning"
  return "unknown"
}

export function buildServices(graph: TopologyGraph, signals: TopologySignals): ServiceView[] {
  const edges = graph.edges ?? []
  const ids = new Set<string>()
  for (const node of graph.nodes ?? []) ids.add(node.id)
  for (const edge of edges) {
    ids.add(edge.source)
    ids.add(edge.target)
  }
  const labels = new Map((graph.nodes ?? []).map((node) => [node.id, node.label || node.id]))

  return [...ids].map((id) => {
    const fleet = signals.fleet.filter((row) => matchesNode(id, row.container, row.service))
    const workloads = signals.workloads.filter((row) => matchesNode(id, row.container, row.service))
    const alerts = signals.alerts.filter((row) => matchesNode(id, row.container))
    const http = signals.http.find((row) => same(id, row.service)) ?? null

    let status: OpStatus = "unknown"
    for (const row of fleet) status = worse(status, fleetStatus(row))
    for (const row of alerts) status = worse(status, alertStatus(row))
    if (http && http.errors > 0) status = worse(status, "warning")
    if (status === "unknown" && http && http.requests > 0 && http.errors === 0) status = "healthy"

    const upstream = edges
      .filter((edge) => edge.target === id)
      .map((edge) => ({ other: edge.source, kind: edge.kind, count: edge.count, origin: edge.origin, port: edge.port }))
      .sort((a, b) => b.count - a.count)
    const downstream = edges
      .filter((edge) => edge.source === id)
      .map((edge) => ({ other: edge.target, kind: edge.kind, count: edge.count, origin: edge.origin, port: edge.port }))
      .sort((a, b) => b.count - a.count)

    return {
      id,
      label: labels.get(id) || id,
      status,
      upstream,
      downstream,
      fleet,
      workloads,
      http,
      alerts,
    }
  })
}

export function rankServices(services: ServiceView[]): ServiceView[] {
  return [...services].sort((a, b) => {
    const byStatus = rank[a.status] - rank[b.status]
    if (byStatus !== 0) return byStatus
    const degree = b.upstream.length + b.downstream.length - (a.upstream.length + a.downstream.length)
    if (degree !== 0) return degree
    return a.label.localeCompare(b.label)
  })
}

export function filterServices(services: ServiceView[], query: string): ServiceView[] {
  const q = query.trim().toLowerCase()
  if (!q) return services
  return services.filter((service) => service.label.toLowerCase().includes(q) || service.id.toLowerCase().includes(q))
}

export function pickService(services: ServiceView[], focus: string): ServiceView | null {
  if (services.length === 0) return null
  return services.find((service) => service.id === focus) ?? services[0]
}

export function busiestWorkload(workloads: WorkloadSnapshot[]): WorkloadSnapshot | null {
  if (workloads.length === 0) return null
  return [...workloads].sort((a, b) => (b.cpu_usage ?? 0) - (a.cpu_usage ?? 0))[0]
}
