import { useCallback, useEffect, useState } from "react"
import { Link } from "react-router-dom"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { fetchInsights, type Insight } from "@/lib/api"
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
}

export function InsightsPage() {
  const [since, setSince] = useState("1h")
  const [insights, setInsights] = useState<Insight[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    setLoading(true)
    fetchInsights(since)
      .then((r) => {
        setInsights(r.insights ?? [])
        setError(null)
      })
      .catch((e) => setError(e instanceof Error ? e.message : "Erro"))
      .finally(() => setLoading(false))
  }, [since])

  useEffect(() => {
    load()
    const t = setInterval(load, 30_000)
    return () => clearInterval(t)
  }, [load])

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Insights</h1>
          <p className="text-muted-foreground text-sm">
            Temas críticos de otimização — memória, GC, OOM e gargalos detectados automaticamente.
          </p>
        </div>
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
            Nenhum tema crítico no período. Continue monitorando memória e logs de GC.
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {insights.map((ins) => (
            <InsightCard key={ins.id} insight={ins} />
          ))}
        </div>
      )}
    </div>
  )
}

function InsightCard({ insight }: { insight: Insight }) {
  const topic =
    insight.theme === "gc_thrashing"
      ? "gc"
      : insight.theme === "oom_risk" || insight.theme === "memory_pressure"
        ? "memory"
        : insight.theme === "error_spike"
          ? "error"
          : "performance"

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
        <div className="flex gap-2 pt-1">
          <Link
            to={`/metrics?container=${encodeURIComponent(insight.container)}`}
            className="border-input bg-background hover:bg-muted inline-flex h-7 items-center rounded-md border px-2.5 text-xs"
          >
            Métricas
          </Link>
          <Link
            to={`/logs?container=${encodeURIComponent(insight.container)}&topic=${topic}`}
            className="border-input bg-background hover:bg-muted inline-flex h-7 items-center rounded-md border px-2.5 text-xs"
          >
            Logs relacionados
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}
