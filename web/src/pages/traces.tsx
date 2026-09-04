import { useCallback, useEffect, useMemo, useState } from "react"
import { Link, useSearchParams } from "react-router-dom"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { fetchTrace, type TraceDetail } from "@/lib/api"

function formatMs(ts: string) {
  return new Date(ts).toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit", second: "2-digit", fractionalSecondDigits: 3 })
}

export function TracesPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [traceId, setTraceId] = useState(searchParams.get("trace_id") ?? "")
  const [detail, setDetail] = useState<TraceDetail | null>(null)
  const [loading, setLoading] = useState(false)

  const load = useCallback((id: string) => {
    const tid = id.trim()
    if (!tid) {
      setDetail(null)
      return
    }
    setLoading(true)
    fetchTrace(tid, "24h")
      .then(setDetail)
      .catch(() => setDetail(null))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    const tid = searchParams.get("trace_id") ?? ""
    setTraceId(tid)
    load(tid)
  }, [searchParams, load])

  const origin = useMemo(() => {
    if (!detail?.spans?.length) return 0
    return detail.spans.reduce((min, sp) => {
      const t = new Date(sp.start_ts).getTime()
      return min === 0 || t < min ? t : min
    }, 0)
  }, [detail])

  const totalWidth = useMemo(() => {
    if (!detail?.duration_ms || detail.duration_ms <= 0) return 1
    return detail.duration_ms
  }, [detail])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Trace explorer</h1>
        <p className="text-muted-foreground text-sm">
          Waterfall por traceId — spans OTLP ou fallback em logs estruturados.
        </p>
      </div>

      <form
        className="flex flex-wrap gap-2"
        onSubmit={(e) => {
          e.preventDefault()
          const tid = traceId.trim()
          if (tid) setSearchParams({ trace_id: tid })
          else setSearchParams({})
        }}
      >
        <Input
          className="max-w-xl font-mono text-sm"
          placeholder="trace_id"
          value={traceId}
          onChange={(e) => setTraceId(e.target.value)}
        />
        <Button type="submit">Buscar</Button>
        <Button type="button" variant="outline" render={<Link to="/logs" />}>
          Logs
        </Button>
      </form>

      {loading ? (
        <Skeleton className="h-48 w-full" />
      ) : !detail || detail.spans.length === 0 ? (
        <Card>
          <CardContent className="text-muted-foreground pt-6 text-sm">
            Nenhum span encontrado para este trace. Tente um traceId de logs recentes ou configure export OTLP para{" "}
            <code className="text-xs">POST /v1/traces</code>.
          </CardContent>
        </Card>
      ) : (
        <>
          <div className="flex flex-wrap items-center gap-2 text-sm">
            <Badge variant="outline">fonte: {detail.source}</Badge>
            <span className="text-muted-foreground font-mono">{detail.trace_id}</span>
            {detail.duration_ms > 0 ? (
              <span className="text-muted-foreground">{detail.duration_ms.toFixed(1)} ms total</span>
            ) : null}
          </div>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Waterfall</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {detail.spans.map((sp) => {
                const startMs = new Date(sp.start_ts).getTime() - origin
                const widthPct = Math.max((sp.duration_ms / totalWidth) * 100, sp.duration_ms > 0 ? 2 : 0.5)
                const leftPct = (startMs / totalWidth) * 100
                const barColor =
                  sp.status === "error" ? "bg-destructive/80" : sp.kind === "log" ? "bg-muted-foreground/50" : "bg-primary/70"
                return (
                  <div key={`${sp.span_id}-${sp.start_ts}`} className="grid gap-1 md:grid-cols-[180px_1fr_120px] md:items-center">
                    <div className="truncate text-xs">
                      <div className="font-medium">{sp.service || sp.container || "—"}</div>
                      <div className="text-muted-foreground truncate">{sp.name}</div>
                    </div>
                    <div className="bg-muted relative h-6 overflow-hidden rounded">
                      <div
                        className={`absolute top-0 h-full rounded ${barColor}`}
                        style={{ left: `${Math.max(leftPct, 0)}%`, width: `${Math.min(widthPct, 100 - leftPct)}%` }}
                        title={`${sp.duration_ms.toFixed(1)} ms`}
                      />
                    </div>
                    <div className="text-muted-foreground flex flex-wrap items-center gap-2 text-xs">
                      <span>{sp.duration_ms > 0 ? `${sp.duration_ms.toFixed(0)} ms` : "—"}</span>
                      <Link
                        className="text-primary hover:underline"
                        to={`/logs?trace_id=${encodeURIComponent(detail.trace_id)}&since=24h`}
                      >
                        logs
                      </Link>
                      {sp.container ? (
                        <Link className="text-primary hover:underline" to={`/logs?container=${encodeURIComponent(sp.container)}`}>
                          svc
                        </Link>
                      ) : null}
                    </div>
                  </div>
                )
              })}
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Timeline</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2 text-xs">
              {detail.spans.map((sp) => (
                <div key={`tl-${sp.span_id}`} className="border-b pb-2 last:border-0">
                  <div className="flex flex-wrap gap-2">
                    <Badge variant="secondary">{sp.kind}</Badge>
                    <span className="font-mono">{formatMs(sp.start_ts)}</span>
                    <span>{sp.service}</span>
                    <span className="text-muted-foreground">{sp.name}</span>
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
        </>
      )}
    </div>
  )
}
