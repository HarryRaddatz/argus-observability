import { useCallback, useEffect, useState } from "react"
import { Link } from "react-router-dom"

import { TimeRangePicker } from "@/components/filters/time-range-picker"
import { PageHeader } from "@/components/layout/page-header"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useQueryState } from "@/hooks/use-query-state"
import { fetchTopology, type TopologyGraph } from "@/lib/api"

const kindLabel: Record<string, string> = { http: "HTTP", amqp: "Fila AMQP" }

function eventsLabel(n: number) {
  return n === 1 ? "1 chamada" : `${n} chamadas`
}

export function TopologyPage() {
  const [since, setSince] = useQueryState("since", "24h")
  const [graph, setGraph] = useState<TopologyGraph>({ nodes: [], edges: [] })
  const [loading, setLoading] = useState(true)

  const load = useCallback(() => {
    setLoading(true)
    fetchTopology(since)
      .then(setGraph)
      .catch(() => setGraph({ nodes: [], edges: [] }))
      .finally(() => setLoading(false))
  }, [since])

  useEffect(() => {
    load()
  }, [load])

  const edges = [...graph.edges].sort((a, b) => b.count - a.count)

  return (
    <div className="space-y-6">
      <PageHeader
        title="Topologia"
        description="Chamadas entre serviços encontradas nos logs."
        actions={<TimeRangePicker value={since} onChange={setSince} />}
      />

      {loading ? (
        <Skeleton className="h-48 w-full" />
      ) : edges.length === 0 ? (
        <p className="text-muted-foreground text-sm">
          Nenhuma dependência encontrada no período. Ela aparece quando os logs registram chamadas HTTP ou AMQP entre serviços.
        </p>
      ) : (
        <>
          <section className="space-y-2">
            <h2 className="text-sm font-medium">Serviços</h2>
            <div className="flex flex-wrap gap-2">
              {graph.nodes.map((n) => (
                <Link key={n.id} to={`/logs?container=${encodeURIComponent(n.id)}&since=${since}`}>
                  <Badge variant="outline">{n.label}</Badge>
                </Link>
              ))}
            </div>
          </section>
          <section className="space-y-2">
            <h2 className="text-sm font-medium">Dependências</h2>
            <div className="grid gap-3 md:grid-cols-2">
              {edges.map((e) => (
                <Card key={`${e.source}-${e.target}-${e.kind}`}>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium">
                      {e.source} → {e.target}
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-2 text-xs">
                    <p className="text-sm font-medium tabular-nums">{eventsLabel(e.count)}</p>
                    <p className="text-muted-foreground">{kindLabel[e.kind] ?? e.kind}</p>
                    <div className="flex flex-wrap gap-3">
                      <Link
                        to={`/logs?container=${encodeURIComponent(e.target)}&topic=error&since=${since}`}
                        className="text-primary hover:underline"
                      >
                        Erros em {e.target}
                      </Link>
                      <Link
                        to={`/traces?service=${encodeURIComponent(e.target)}&since=${since}`}
                        className="text-primary hover:underline"
                      >
                        Traces de {e.target}
                      </Link>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </section>
        </>
      )}
    </div>
  )
}
