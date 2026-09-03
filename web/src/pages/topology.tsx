import { useCallback, useEffect, useState } from "react"
import { Link } from "react-router-dom"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { fetchTopology, type TopologyGraph } from "@/lib/api"
import { TIME_RANGES } from "@/lib/observability"

export function TopologyPage() {
  const [since, setSince] = useState("24h")
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

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Service map</h1>
          <p className="text-muted-foreground text-sm">
            Topologia inferida de logs HTTP e AMQP entre serviços.
          </p>
        </div>
        <div className="flex gap-1">
          {TIME_RANGES.map((r) => (
            <Button
              key={r.id}
              size="sm"
              variant={since === r.id ? "default" : "outline"}
              onClick={() => setSince(r.id)}
            >
              {r.label}
            </Button>
          ))}
        </div>
      </div>

      {loading ? (
        <Skeleton className="h-48 w-full" />
      ) : graph.edges.length === 0 ? (
        <Card>
          <CardContent className="text-muted-foreground pt-6 text-sm">
            Nenhuma aresta inferida ainda — aguarde ingestão de logs com HTTP/AMQP.
          </CardContent>
        </Card>
      ) : (
        <>
          <div className="flex flex-wrap gap-2">
            {graph.nodes.map((n) => (
              <Link key={n.id} to={`/logs?container=${encodeURIComponent(n.id)}`}>
                <Badge variant="outline">{n.label}</Badge>
              </Link>
            ))}
          </div>
          <div className="grid gap-3 md:grid-cols-2">
            {graph.edges.map((e) => (
              <Card key={`${e.source}-${e.target}-${e.kind}`}>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm font-medium">
                    {e.source} → {e.target}
                  </CardTitle>
                </CardHeader>
                <CardContent className="text-muted-foreground flex justify-between text-xs">
                  <span>{e.kind}</span>
                  <span>{e.count} eventos</span>
                  <Link
                    to={`/logs?container=${encodeURIComponent(e.target)}&topic=error`}
                    className="text-primary hover:underline"
                  >
                    Logs destino
                  </Link>
                </CardContent>
              </Card>
            ))}
          </div>
        </>
      )}
    </div>
  )
}
