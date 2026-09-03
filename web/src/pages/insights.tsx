import { useCallback, useEffect, useState } from "react"
import { Link, useSearchParams } from "react-router-dom"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { fetchInsights, listWorkloadGroups, type Insight, type WorkloadGroup } from "@/lib/api"
import { TIME_RANGES } from "@/lib/observability"

const severityVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  critical: "destructive",
  warning: "secondary",
  info: "outline",
}

const themeLabels: Record<string, string> = {
  memory_pressure: "Memória",
  gc_thrashing: "GC",
  oom_risk: "OOM",
  error_spike: "Erros",
  cpu_hot: "CPU",
  restart_loop: "Restarts",
  oom_killed: "OOM kill",
  unhealthy: "Health",
  group_degradation: "Grupo",
  alert_active: "Alerta",
  log_pattern_spike: "Pattern",
  chain_degradation: "Cadeia",
}

export function InsightsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [since, setSince] = useState(searchParams.get("since") ?? "1h")
  const [group, setGroup] = useState(searchParams.get("group") ?? "")
  const [groups, setGroups] = useState<WorkloadGroup[]>([])
  const [insights, setInsights] = useState<Insight[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    listWorkloadGroups().then(setGroups).catch(() => setGroups([]))
  }, [])

  const load = useCallback(() => {
    setLoading(true)
    fetchInsights(since, group || undefined)
      .then((r) => {
        setInsights(r.insights ?? [])
        setError(null)
      })
      .catch((e) => setError(e instanceof Error ? e.message : "Erro"))
      .finally(() => setLoading(false))
  }, [since, group])

  useEffect(() => {
    load()
    const t = setInterval(load, 30_000)
    return () => clearInterval(t)
  }, [load])

  useEffect(() => {
    const p = new URLSearchParams()
    if (since !== "1h") p.set("since", since)
    if (group) p.set("group", group)
    setSearchParams(p, { replace: true })
  }, [since, group, setSearchParams])

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Insights</h1>
          <p className="text-muted-foreground text-sm">
            Temas agregados por container ou grupo — memória, GC, alertas e padrões.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <select
            className="border-input bg-background h-9 rounded-md border px-3 text-sm"
            value={group}
            onChange={(e) => setGroup(e.target.value)}
          >
            <option value="">Todos containers</option>
            {groups.map((g) => (
              <option key={g.id} value={g.id}>
                {g.name}
              </option>
            ))}
          </select>
          <div className="flex gap-1 rounded-md border p-1">
            {TIME_RANGES.map((r) => (
              <Button
                key={r.id}
                size="sm"
                variant={since === r.id ? "default" : "ghost"}
                onClick={() => setSince(r.id)}
              >
                {r.label}
              </Button>
            ))}
          </div>
        </div>
      </div>

      {error ? <p className="text-destructive text-sm">{error}</p> : null}

      {loading && insights.length === 0 ? (
        <div className="grid gap-4 md:grid-cols-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-40 w-full" />
          ))}
        </div>
      ) : insights.length === 0 ? (
        <Card>
          <CardContent className="text-muted-foreground pt-6 text-sm">
            Nenhum tema crítico no período.
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {insights.map((ins) => (
            <InsightCard key={ins.id} insight={ins} group={group} />
          ))}
        </div>
      )}
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

  return (
    <Card className={insight.severity === "critical" ? "border-destructive/40" : undefined}>
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between gap-2">
          <CardTitle className="text-base leading-snug">{insight.title}</CardTitle>
          <Badge variant={severityVariant[insight.severity] ?? "outline"}>{insight.severity}</Badge>
        </div>
        <CardDescription className="flex gap-2">
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
          {!group ? (
            <Link
              to={`/metrics?container=${encodeURIComponent(insight.container)}`}
              className="border-input bg-background hover:bg-muted inline-flex h-7 items-center rounded-md border px-2.5 text-xs"
            >
              Métricas
            </Link>
          ) : (
            <Link
              to={`/explorer?group=${encodeURIComponent(group)}`}
              className="border-input bg-background hover:bg-muted inline-flex h-7 items-center rounded-md border px-2.5 text-xs"
            >
              Explorer grupo
            </Link>
          )}
          <Link
            to={logsLink}
            className="border-input bg-background hover:bg-muted inline-flex h-7 items-center rounded-md border px-2.5 text-xs"
          >
            Logs relacionados
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}
