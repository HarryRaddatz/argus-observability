import { useCallback, useEffect, useRef, useState } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"

import { ListPager } from "@/components/list-pager"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { usePolling } from "@/hooks/use-polling"
import { useQueryState } from "@/hooks/use-query-state"
import { fetchLogPatterns, type LogPattern, type LogPatternParams } from "@/lib/api"
import { PAGE_SIZE, pageIndex, pageSlice } from "@/lib/page"

type Props = {
  filters: LogPatternParams
  onOpenLines: (pattern: LogPattern) => void
}

export function LogPatterns({ filters, onOpenLines }: Props) {
  const { t } = useTranslation()
  const [pageRaw, setPage] = useQueryState("page", "1")
  const [rows, setRows] = useState<LogPattern[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const { since, q, container, group } = filters
  const filterKey = `${since}|${q}|${container}|${group}`
  const filterSeen = useRef(filterKey)
  useEffect(() => {
    if (filterSeen.current === filterKey) return
    filterSeen.current = filterKey
    setPage("1")
  }, [filterKey, setPage])

  const load = useCallback(() => {
    fetchLogPatterns({ since, q, container, group })
      .then((r) => {
        setRows(r)
        setError(null)
      })
      .catch((e) => setError(e instanceof Error ? e.message : t("logs.patternsError")))
      .finally(() => setLoading(false))
  }, [since, q, container, group, t])

  usePolling(load)

  if (loading && rows.length === 0) return <Skeleton className="h-24 w-full" />
  if (error) return <p className="text-destructive text-sm">{error}</p>
  if (rows.length === 0) {
    return <p className="text-muted-foreground text-sm">{t("logs.patternsEmpty")}</p>
  }

  const patternPage = pageSlice(rows, pageIndex(pageRaw), PAGE_SIZE)
  return (
    <div className="space-y-3">
    <div className="grid gap-3 lg:grid-cols-2">
      {patternPage.rows.map((p) => (
        <Card key={`${p.pattern_key}-${p.container}`}>
          <CardHeader className="pb-2">
            <div className="flex items-start justify-between gap-2">
              <CardTitle className="min-w-0 truncate text-sm font-medium">{p.service || p.container}</CardTitle>
              <Badge variant={p.count >= 200 ? "destructive" : p.count >= 50 ? "secondary" : "outline"}>
                {p.count === 1 ? t("logs.oneOccurrence") : t("logs.nOccurrences", { count: p.count })}
              </Badge>
            </div>
            <Link
              to={`/metrics?container=${encodeURIComponent(p.container)}`}
              className="text-muted-foreground truncate text-xs hover:underline"
              title={t("logs.metricsOf", { container: p.container })}
            >
              {p.container}
            </Link>
          </CardHeader>
          <CardContent className="space-y-2 text-xs">
            <p className="font-mono break-all">{p.pattern}</p>
            <Button size="sm" variant="outline" onClick={() => onOpenLines(p)}>
              {t("logs.viewLines")}
            </Button>
          </CardContent>
        </Card>
      ))}
    </div>
    <ListPager
      start={patternPage.start}
      end={patternPage.end}
      total={patternPage.total}
      page={patternPage.page}
      pageSize={PAGE_SIZE}
      onPage={(n) => setPage(String(n))}
    />
    </div>
  )
}
