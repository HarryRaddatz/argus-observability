import { useCallback, useEffect, useState } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"

import { TimeRangePicker } from "@/components/filters/time-range-picker"
import { PageHeader } from "@/components/layout/page-header"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useQueryState } from "@/hooks/use-query-state"
import { fetchTopology, type TopologyGraph } from "@/lib/api"

export function TopologyPage() {
  const { t } = useTranslation()
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
        title={t("topology.title")}
        description={t("topology.description")}
        actions={<TimeRangePicker value={since} onChange={setSince} />}
      />

      {loading ? (
        <Skeleton className="h-48 w-full" />
      ) : edges.length === 0 ? (
        <p className="text-muted-foreground text-sm">
          {t("topology.empty")}
        </p>
      ) : (
        <>
          <section className="space-y-2">
            <h2 className="text-sm font-medium">{t("topology.services")}</h2>
            <div className="flex flex-wrap gap-2">
              {graph.nodes.map((n) => (
                <Link key={n.id} to={`/logs?container=${encodeURIComponent(n.id)}&since=${since}`}>
                  <Badge variant="outline">{n.label}</Badge>
                </Link>
              ))}
            </div>
          </section>
          <section className="space-y-2">
            <h2 className="text-sm font-medium">{t("topology.deps")}</h2>
            <div className="grid gap-3 md:grid-cols-2">
              {edges.map((e) => (
                <Card key={`${e.source}-${e.target}-${e.kind}`}>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium">
                      {e.source} → {e.target}
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-2 text-xs">
                    <p className="text-sm font-medium tabular-nums">
                      {e.count === 1 ? t("topology.oneCall") : t("topology.nCalls", { count: e.count })}
                    </p>
                    <p className="text-muted-foreground">{e.kind === "amqp" ? t("topology.amqp") : e.kind === "http" ? t("topology.http") : e.kind}</p>
                    <div className="flex flex-wrap gap-3">
                      <Link
                        to={`/logs?container=${encodeURIComponent(e.target)}&topic=error&since=${since}`}
                        className="text-primary hover:underline"
                      >
                        {t("topology.errorsIn", { name: e.target })}
                      </Link>
                      <Link
                        to={`/traces?service=${encodeURIComponent(e.target)}&since=${since}`}
                        className="text-primary hover:underline"
                      >
                        {t("topology.tracesOf", { name: e.target })}
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
