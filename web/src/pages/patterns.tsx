import { useCallback, useEffect, useState } from "react"
import { Link } from "react-router-dom"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import { fetchLogPatterns, type LogPattern } from "@/lib/api"
import { TIME_RANGES } from "@/lib/observability"

export function PatternsPage() {
  const [since, setSince] = useState("1h")
  const [rows, setRows] = useState<LogPattern[]>([])
  const [loading, setLoading] = useState(true)

  const load = useCallback(() => {
    setLoading(true)
    fetchLogPatterns(since)
      .then(setRows)
      .catch(() => setRows([]))
      .finally(() => setLoading(false))
  }, [since])

  useEffect(() => {
    load()
  }, [load])

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Log patterns</h1>
          <p className="text-muted-foreground text-sm">
            Mensagens normalizadas agrupadas — detecta repetição e spikes.
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

      <ScrollArea className="h-[560px] rounded-md border">
        <div className="space-y-3 p-4">
          {loading && rows.length === 0 ? (
            <Skeleton className="h-24 w-full" />
          ) : rows.length === 0 ? (
            <p className="text-muted-foreground text-sm">Nenhum pattern no período.</p>
          ) : (
            rows.map((p) => (
              <Card key={`${p.pattern_key}-${p.container}`}>
                <CardHeader className="pb-2">
                  <div className="flex items-center justify-between gap-2">
                    <CardTitle className="text-sm font-medium">{p.service || p.container}</CardTitle>
                    <Badge variant={p.count >= 200 ? "destructive" : p.count >= 50 ? "secondary" : "outline"}>
                      {p.count}x
                    </Badge>
                  </div>
                </CardHeader>
                <CardContent className="space-y-2 text-xs">
                  <p className="font-mono break-all">{p.pattern}</p>
                  <div className="flex gap-2">
                    <Link
                      to={`/logs?container=${encodeURIComponent(p.container)}&q=${encodeURIComponent(p.pattern.slice(0, 40))}`}
                      className="text-primary hover:underline"
                    >
                      Ver logs
                    </Link>
                  </div>
                </CardContent>
              </Card>
            ))
          )}
        </div>
      </ScrollArea>
    </div>
  )
}
