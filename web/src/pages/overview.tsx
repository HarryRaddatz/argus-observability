import { useCallback, useMemo, useState, type ReactNode } from "react"
import { Link } from "react-router-dom"

import { PageHeader } from "@/components/layout/page-header"
import { StatCard } from "@/components/metrics/stat-card"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
import { usePolling } from "@/hooks/use-polling"
import {
  fetchActiveAlerts,
  fetchFleetStatus,
  fetchHTTPSummary,
  fetchSLOStatuses,
  getHealth,
  type ActiveAlert,
  type ContainerFleetStatus,
  type HTTPServiceSummary,
  type SLOStatus,
} from "@/lib/api"
import { instabilityReason, isUnstable } from "@/lib/container-state"
import { formatPercent } from "@/lib/format"
import { severityLabel, severityVariant } from "@/lib/severity"
import { sloMetricsLink } from "@/lib/slo"

const HTTP_ERROR_RATE = 0.01
const LIST_LIMIT = 8

type State = {
  alerts: ActiveAlert[]
  slos: SLOStatus[]
  unstable: ContainerFleetStatus[]
  http: HTTPServiceSummary[]
}

type ActionItem = {
  key: string
  title: string
  detail: string
  status: ReactNode
  to: string
}

const EMPTY: State = { alerts: [], slos: [], unstable: [], http: [] }

export function OverviewPage() {
  const [data, setData] = useState<State>(EMPTY)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    getHealth()
      .then(() => Promise.all([fetchActiveAlerts(), fetchSLOStatuses(), fetchFleetStatus(), fetchHTTPSummary("1h")]))
      .then(([alerts, slos, fleet, http]) => {
        setData({
          alerts,
          slos: slos.filter((s) => s.breached || s.error_budget_remaining < 30),
          unstable: fleet.containers.filter(isUnstable),
          http: http.filter((s) => s.requests > 0 && s.error_rate >= HTTP_ERROR_RATE).sort((a, b) => b.error_rate - a.error_rate),
        })
        setError(null)
      })
      .catch(() => setError("O hub não respondeu. Os dados abaixo podem estar desatualizados."))
      .finally(() => setLoading(false))
  }, [])

  usePolling(load)

  const breached = data.slos.filter((s) => s.breached).length

  const items = useMemo<ActionItem[]>(() => {
    const list: ActionItem[] = []
    for (const a of data.alerts) {
      list.push({
        key: `alert-${a.rule_id}-${a.entity_uid}`,
        title: a.title,
        detail: a.container,
        status: <Badge variant={severityVariant[a.severity] ?? "outline"}>{severityLabel[a.severity] ?? a.severity}</Badge>,
        to: a.container ? `/logs?container=${encodeURIComponent(a.container)}` : "/problems",
      })
    }
    for (const s of data.slos) {
      list.push({
        key: `slo-${s.slo.id}`,
        title: s.slo.name,
        detail: `SLO de ${s.slo.service}`,
        status: <Badge variant={s.breached ? "destructive" : "secondary"}>{s.breached ? "Violado" : "Margem baixa"}</Badge>,
        to: sloMetricsLink(s.slo),
      })
    }
    for (const c of data.unstable) {
      list.push({
        key: `container-${c.entity_uid}`,
        title: c.container,
        detail: instabilityReason(c),
        status: <Badge variant="destructive">Container</Badge>,
        to: `/logs?container=${encodeURIComponent(c.container)}`,
      })
    }
    for (const s of data.http) {
      list.push({
        key: `http-${s.service}`,
        title: s.service,
        detail: `${formatPercent(s.error_rate * 100)} das requisições HTTP com erro`,
        status: <Badge variant={s.error_rate >= 0.05 ? "destructive" : "secondary"}>HTTP</Badge>,
        to: `/metrics?mode=compare&metric=http.error_rate&service=${encodeURIComponent(s.service)}`,
      })
    }
    return list
  }, [data])

  return (
    <div className="space-y-8">
      <PageHeader title="Visão geral" />

      {error ? (
        <Card className="border-destructive/50">
          <CardContent className="text-destructive pt-6 text-sm">{error}</CardContent>
        </Card>
      ) : null}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          label="Alertas ativos"
          value={data.alerts.length}
          tone={data.alerts.length > 0 ? "critical" : "default"}
          to="/problems"
          loading={loading}
        />
        <StatCard
          label="SLOs violados"
          value={breached}
          tone={breached > 0 ? "critical" : data.slos.length > 0 ? "warning" : "default"}
          hint={data.slos.length > breached ? `${data.slos.length - breached} com margem baixa` : undefined}
          to="/problems?tab=slos"
          loading={loading}
        />
        <StatCard
          label="Containers com problema"
          value={data.unstable.length}
          tone={data.unstable.length > 0 ? "critical" : "default"}
          to="/containers?filter=unstable&view=table"
          loading={loading}
        />
        <StatCard
          label="Serviços HTTP com erro"
          value={data.http.length}
          tone={data.http.length > 0 ? "warning" : "default"}
          hint="Mais de 1% em 1 hora"
          to="/metrics?mode=compare&metric=http.error_rate"
          loading={loading}
        />
      </div>

      {loading ? null : items.length === 0 ? (
        error ? null : <p className="text-muted-foreground text-sm">Nada pede ação agora.</p>
      ) : (
        <section className="space-y-3">
          <h2 className="text-lg font-medium">Pedem ação agora</h2>
          <ul className="divide-y rounded-md border">
            {items.slice(0, LIST_LIMIT).map((item) => (
              <li key={item.key}>
                <Link
                  to={item.to}
                  className="hover:bg-muted/50 focus-visible:ring-ring flex items-center justify-between gap-3 px-3 py-2 focus-visible:ring-2 focus-visible:outline-none"
                >
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">{item.title}</p>
                    <p className="text-muted-foreground truncate text-xs">{item.detail}</p>
                  </div>
                  <div className="shrink-0">{item.status}</div>
                </Link>
              </li>
            ))}
          </ul>
          {items.length > LIST_LIMIT ? (
            <p className="text-muted-foreground text-sm">
              Mais {items.length - LIST_LIMIT} nos indicadores acima.
            </p>
          ) : null}
        </section>
      )}
    </div>
  )
}
