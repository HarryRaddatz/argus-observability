import { useCallback, useState } from "react"
import { Link } from "react-router-dom"

import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { usePolling } from "@/hooks/use-polling"
import { fetchActiveAlerts, fetchInsights, type ActiveAlert, type Insight } from "@/lib/api"
import { severityLabel, severityVariant } from "@/lib/severity"

const themeLabels: Record<string, string> = {
  memory_pressure: "Memória",
  gc_thrashing: "GC",
  oom_risk: "Risco de OOM",
  error_spike: "Erros",
  cpu_hot: "CPU",
  restart_loop: "Reinícios",
  oom_killed: "OOM",
  unhealthy: "Healthcheck",
  group_degradation: "Grupo",
  alert_active: "Alerta",
  log_pattern_spike: "Padrão de log",
  chain_degradation: "Cadeia de serviços",
}

const linkClass =
  "border-input bg-background hover:bg-muted inline-flex h-7 items-center rounded-md border px-2.5 text-xs"

type Props = {
  since: string
  group: string
}

export function ProblemsNow({ since, group }: Props) {
  const [alerts, setAlerts] = useState<ActiveAlert[]>([])
  const [insights, setInsights] = useState<Insight[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    Promise.all([fetchActiveAlerts(), fetchInsights(since, group || undefined)])
      .then(([a, i]) => {
        setAlerts(a)
        setInsights(i.insights ?? [])
        setError(null)
      })
      .catch((e) => setError(e instanceof Error ? e.message : "Não foi possível carregar os problemas"))
      .finally(() => setLoading(false))
  }, [since, group])

  usePolling(load)

  if (loading && alerts.length === 0 && insights.length === 0) {
    return (
      <div className="grid gap-4 md:grid-cols-2">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-40 w-full" />
        ))}
      </div>
    )
  }

  return (
    <div className="space-y-8">
      {error ? <p className="text-destructive text-sm">{error}</p> : null}

      <section className="space-y-3">
        <h2 className="text-lg font-medium">Alertas ativos</h2>
        {alerts.length === 0 ? (
          <p className="text-muted-foreground text-sm">Nenhuma regra de alerta disparada agora.</p>
        ) : (
          <div className="grid gap-3 md:grid-cols-2">
            {alerts.map((a) => (
              <Card key={a.rule_id + a.entity_uid} className="border-destructive/30">
                <CardHeader className="pb-2">
                  <div className="flex items-start justify-between gap-2">
                    <CardTitle className="text-sm">{a.title}</CardTitle>
                    <Badge variant={severityVariant[a.severity] ?? "outline"}>{severityLabel[a.severity] ?? a.severity}</Badge>
                  </div>
                  <CardDescription>
                    {a.container}, desde {new Date(a.fired_at).toLocaleString("pt-BR")}
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-3 text-sm">
                  <p className="text-muted-foreground">{a.summary}</p>
                  {a.container ? (
                    <div className="flex flex-wrap gap-2">
                      <Link to={`/logs?container=${encodeURIComponent(a.container)}`} className={linkClass}>
                        Logs do container
                      </Link>
                      <Link to={`/metrics?container=${encodeURIComponent(a.container)}`} className={linkClass}>
                        Métricas do container
                      </Link>
                    </div>
                  ) : null}
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </section>

      <section className="space-y-3">
        <h2 className="text-lg font-medium">Insights</h2>
        {insights.length === 0 ? (
          <p className="text-muted-foreground text-sm">Nenhum achado automático no período.</p>
        ) : (
          <div className="grid gap-4 md:grid-cols-2">
            {insights.map((ins) => (
              <InsightCard key={ins.id} insight={ins} group={group} />
            ))}
          </div>
        )}
      </section>
    </div>
  )
}

function InsightCard({ insight, group }: { insight: Insight; group: string }) {
  const topic =
    insight.theme === "gc_thrashing"
      ? "gc"
      : insight.theme === "oom_risk" || insight.theme === "memory_pressure"
        ? "memory"
        : insight.theme === "error_spike"
          ? "error"
          : "performance"

  const logsLink = group
    ? `/logs?group=${encodeURIComponent(group)}&topic=${topic}`
    : `/logs?container=${encodeURIComponent(insight.container)}&topic=${topic}`
  const metricsLink = group
    ? `/metrics?mode=compare&group=${encodeURIComponent(group)}`
    : `/metrics?container=${encodeURIComponent(insight.container)}`

  return (
    <Card className={insight.severity === "critical" ? "border-destructive/40" : undefined}>
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between gap-2">
          <CardTitle className="text-base leading-snug">{insight.title}</CardTitle>
          <Badge variant={severityVariant[insight.severity] ?? "outline"}>{severityLabel[insight.severity] ?? insight.severity}</Badge>
        </div>
        <CardDescription className="flex flex-wrap items-center gap-2">
          <Badge variant="outline">{themeLabels[insight.theme] ?? insight.theme}</Badge>
          <span>{insight.container}</span>
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3 text-sm">
        <p>{insight.summary}</p>
        {insight.recommendations?.length ? (
          <ul className="text-muted-foreground list-inside list-disc space-y-1">
            {insight.recommendations.map((r) => (
              <li key={r}>{r}</li>
            ))}
          </ul>
        ) : null}
        <div className="flex flex-wrap gap-2 pt-1">
          <Link to={metricsLink} className={linkClass}>
            {group ? "Métricas do grupo" : "Métricas do container"}
          </Link>
          <Link to={logsLink} className={linkClass}>
            Logs relacionados
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}
