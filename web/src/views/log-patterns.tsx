import { useCallback, useState } from "react"
import { Link } from "react-router-dom"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { usePolling } from "@/hooks/use-polling"
import { fetchLogPatterns, type LogPattern, type LogPatternParams } from "@/lib/api"

type Props = {
  filters: LogPatternParams
  onOpenLines: (pattern: LogPattern) => void
}

function countLabel(n: number) {
  return n === 1 ? "1 ocorrência" : `${n} ocorrências`
}

export function LogPatterns({ filters, onOpenLines }: Props) {
  const [rows, setRows] = useState<LogPattern[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const { since, q, container, group } = filters

  const load = useCallback(() => {
    fetchLogPatterns({ since, q, container, group })
      .then((r) => {
        setRows(r)
        setError(null)
      })
      .catch((e) => setError(e instanceof Error ? e.message : "Não foi possível carregar os padrões"))
      .finally(() => setLoading(false))
  }, [since, q, container, group])

  usePolling(load)

  if (loading && rows.length === 0) return <Skeleton className="h-24 w-full" />
  if (error) return <p className="text-destructive text-sm">{error}</p>
  if (rows.length === 0) {
    return <p className="text-muted-foreground text-sm">Nenhum padrão repetido com estes filtros no período.</p>
  }

  return (
    <div className="grid gap-3 lg:grid-cols-2">
      {rows.map((p) => (
        <Card key={`${p.pattern_key}-${p.container}`}>
          <CardHeader className="pb-2">
            <div className="flex items-start justify-between gap-2">
              <CardTitle className="min-w-0 truncate text-sm font-medium">{p.service || p.container}</CardTitle>
              <Badge variant={p.count >= 200 ? "destructive" : p.count >= 50 ? "secondary" : "outline"}>{countLabel(p.count)}</Badge>
            </div>
            <Link
              to={`/metrics?container=${encodeURIComponent(p.container)}`}
              className="text-muted-foreground truncate text-xs hover:underline"
              title={`Métricas de ${p.container}`}
            >
              {p.container}
            </Link>
          </CardHeader>
          <CardContent className="space-y-2 text-xs">
            <p className="font-mono break-all">{p.pattern}</p>
            <Button size="sm" variant="outline" onClick={() => onOpenLines(p)}>
              Ver as linhas
            </Button>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
