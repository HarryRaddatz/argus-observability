import { useCallback, useEffect, useState } from "react"
import { Link } from "react-router-dom"

import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { fetchSLOStatuses, type SLOStatus } from "@/lib/api"

function Bar({ value }: { value: number }) {
  return (
    <div className="bg-muted h-2 overflow-hidden rounded">
      <div className="bg-primary h-full rounded" style={{ width: `${Math.min(Math.max(value, 0), 100)}%` }} />
    </div>
  )
}

function budgetVariant(pct: number) {
  if (pct < 10) return "destructive" as const
  if (pct < 30) return "secondary" as const
  return "outline" as const
}

export function SLOsPage() {
  const [rows, setRows] = useState<SLOStatus[]>([])
  const [loading, setLoading] = useState(true)

  const load = useCallback(() => {
    setLoading(true)
    fetchSLOStatuses()
      .then(setRows)
      .catch(() => setRows([]))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    load()
  }, [load])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">SLOs</h1>
        <p className="text-muted-foreground text-sm">
          Objetivos de nível de serviço por serviço — compliance, p95 e error budget.
        </p>
      </div>

      {loading ? (
        <Skeleton className="h-48 w-full" />
      ) : rows.length === 0 ? (
        <Card>
          <CardContent className="text-muted-foreground pt-6 text-sm">
            Nenhum SLO configurado. O seed padrão inclui agendamentoapi após migração do hub.
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {rows.map((row) => (
            <Card key={row.slo.id}>
              <CardHeader className="pb-2">
                <div className="flex items-start justify-between gap-2">
                  <CardTitle className="text-base font-medium">{row.slo.name}</CardTitle>
                  <Badge variant={row.breached ? "destructive" : "outline"}>
                    {row.breached ? "violação" : "ok"}
                  </Badge>
                </div>
                <p className="text-muted-foreground text-xs">
                  {row.slo.service}
                  {row.slo.sli_metric === "latency_p95"
                    ? ` · p95 < ${row.slo.latency_threshold_ms} ms`
                    : " · disponibilidade"}
                  {" · "}
                  target {row.slo.target}%
                </p>
              </CardHeader>
              <CardContent className="space-y-3 text-sm">
                <div>
                  <div className="mb-1 flex justify-between text-xs">
                    <span>Compliance</span>
                    <span>{row.compliance.toFixed(2)}%</span>
                  </div>
                  <Bar value={row.compliance} />
                </div>
                <div>
                  <div className="mb-1 flex justify-between text-xs">
                    <span>Error budget restante</span>
                    <Badge variant={budgetVariant(row.error_budget_remaining)}>
                      {row.error_budget_remaining.toFixed(1)}%
                    </Badge>
                  </div>
                  <Bar value={row.error_budget_remaining} />
                </div>
                <div className="text-muted-foreground grid grid-cols-2 gap-2 text-xs">
                  <span>Eventos: {row.total_events}</span>
                  <span>Bons: {row.good_events}</span>
                  {row.p95_latency_ms > 0 ? <span>p95: {row.p95_latency_ms.toFixed(0)} ms</span> : null}
                  <span>Janela: {row.slo.window_hours}h</span>
                </div>
                {row.slo.service ? (
                  <Link className="text-primary text-xs hover:underline" to={`/metrics?service=${encodeURIComponent(row.slo.service)}`}>
                    Ver métricas HTTP
                  </Link>
                ) : null}
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
