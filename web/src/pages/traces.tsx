import { useCallback, useEffect, useMemo, useState } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"

import { TimeRangePicker } from "@/components/filters/time-range-picker"
import { PageHeader } from "@/components/layout/page-header"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useQueryPatch, useQueryState } from "@/hooks/use-query-state"
import { fetchTrace, fetchTraces, type TraceDetail, type TraceSummary } from "@/lib/api"
import { localeBcp47 } from "@/i18n"

function formatClock(ts: string, lng: string) {
  return new Date(ts).toLocaleTimeString(localeBcp47(lng), {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    fractionalSecondDigits: 3,
  })
}

function formatDuration(ms: number) {
  if (ms <= 0) return "—"
  if (ms < 1000) return `${ms.toFixed(0)} ms`
  return `${(ms / 1000).toFixed(2)} s`
}

export function TracesPage() {
  const { t } = useTranslation()
  const [traceId] = useQueryState("trace_id", "")
  const [since, setSince] = useQueryState("since", "1h")
  const [service, setService] = useQueryState("service", "")
  const patch = useQueryPatch()
  const [draft, setDraft] = useState(traceId)
  const [seenId, setSeenId] = useState(traceId)
  if (traceId !== seenId) {
    setSeenId(traceId)
    setDraft(traceId)
  }

  return (
    <div className="space-y-6">
      <PageHeader title={t("traces.title")} actions={<TimeRangePicker value={since} onChange={setSince} />} />

      <form
        role="search"
        className="flex flex-wrap gap-2"
        onSubmit={(e) => {
          e.preventDefault()
          patch({ trace_id: draft.trim() })
        }}
      >
        <Input
          aria-label="Trace ID"
          className="max-w-xl font-mono text-sm"
          placeholder={t("traces.openById")}
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
        />
        <Button type="submit" disabled={!draft.trim()}>
          {t("traces.open")}
        </Button>
      </form>

      {traceId ? (
        <TraceDetailView traceId={traceId} onBack={() => patch({ trace_id: null })} />
      ) : (
        <TraceList since={since} service={service} onService={setService} onOpen={(id) => patch({ trace_id: id })} />
      )}
    </div>
  )
}

type ListProps = {
  since: string
  service: string
  onService: (service: string) => void
  onOpen: (traceId: string) => void
}

function TraceList({ since, service, onService, onOpen }: ListProps) {
  const { t, i18n } = useTranslation()
  const [rows, setRows] = useState<TraceSummary[]>([])
  const [services, setServices] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    setLoading(true)
    fetchTraces(since, service || undefined)
      .then((r) => {
        setRows(r)
        setError(null)
        if (!service) setServices([...new Set(r.map((row) => row.service).filter(Boolean))].sort())
      })
      .catch((e) => setError(e instanceof Error ? e.message : t("traces.loadError")))
      .finally(() => setLoading(false))
  }, [since, service, t])

  useEffect(() => {
    load()
  }, [load])

  const serviceOptions = useMemo(
    () => (service && !services.includes(service) ? [service, ...services] : services),
    [service, services],
  )

  return (
    <section className="space-y-3">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-lg font-medium">{t("traces.recent")}</h2>
        <select
          aria-label={t("traces.service")}
          className="border-input bg-background h-8 rounded-md border px-2 text-sm"
          value={service}
          onChange={(e) => onService(e.target.value)}
        >
          <option value="">{t("traces.allServices")}</option>
          {serviceOptions.map((s) => (
            <option key={s} value={s}>
              {s}
            </option>
          ))}
        </select>
      </div>
      {error ? <p className="text-destructive text-sm">{error}</p> : null}
      <ScrollArea className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("traces.start")}</TableHead>
              <TableHead>{t("traces.service")}</TableHead>
              <TableHead>{t("traces.operation")}</TableHead>
              <TableHead>{t("traces.duration")}</TableHead>
              <TableHead>{t("traces.spans")}</TableHead>
              <TableHead>{t("traces.outcome")}</TableHead>
              <TableHead>{t("traces.source")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7}>
                  <Skeleton className="h-8 w-full" />
                </TableCell>
              </TableRow>
            ) : rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-muted-foreground text-center">
                  {t("traces.empty")}
                </TableCell>
              </TableRow>
            ) : (
              rows.map((row) => (
                <TableRow key={row.trace_id} className={row.error ? "bg-destructive/5" : undefined}>
                  <TableCell className="whitespace-nowrap font-mono text-xs">{formatClock(row.start_ts, i18n.language)}</TableCell>
                  <TableCell className="text-sm">
                    {row.container ? (
                      <Link
                        to={`/metrics?container=${encodeURIComponent(row.container)}`}
                        className="hover:underline"
                        title={t("traces.metricsOf", { container: row.container })}
                      >
                        {row.service || row.container}
                      </Link>
                    ) : (
                      row.service || "—"
                    )}
                  </TableCell>
                  <TableCell className="max-w-[320px]">
                    <button
                      type="button"
                      className="text-primary block max-w-full truncate text-left text-sm hover:underline"
                      title={row.trace_id}
                      onClick={() => onOpen(row.trace_id)}
                    >
                      {row.name || row.trace_id}
                    </button>
                  </TableCell>
                  <TableCell className="tabular-nums">{formatDuration(row.duration_ms)}</TableCell>
                  <TableCell className="tabular-nums">{row.span_count}</TableCell>
                  <TableCell>
                    {row.error ? <Badge variant="destructive">{t("traces.withError")}</Badge> : <Badge variant="outline">{t("traces.ok")}</Badge>}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-xs">
                    {row.source === "logs" ? t("nav.logs") : row.source.toUpperCase()}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </ScrollArea>
    </section>
  )
}

function TraceDetailView({ traceId, onBack }: { traceId: string; onBack: () => void }) {
  const { t, i18n } = useTranslation()
  const [detail, setDetail] = useState<TraceDetail | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    fetchTrace(traceId, "24h")
      .then((d) => {
        if (!cancelled) setDetail(d)
      })
      .catch(() => {
        if (!cancelled) setDetail(null)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [traceId])

  const origin = useMemo(() => {
    if (!detail?.spans?.length) return 0
    return Math.min(...detail.spans.map((sp) => new Date(sp.start_ts).getTime()))
  }, [detail])

  const totalWidth = detail?.duration_ms && detail.duration_ms > 0 ? detail.duration_ms : 1

  const sourceLabel = (source: string) => (source === "logs" ? t("nav.logs") : source.toUpperCase())

  return (
    <section className="space-y-4">
      <div className="flex flex-wrap items-center gap-3">
        <Button size="sm" variant="outline" onClick={onBack}>
          {t("traces.backToList")}
        </Button>
        <Link to={`/logs?trace_id=${encodeURIComponent(traceId)}&since=24h`} className="text-primary text-sm hover:underline">
          {t("traces.logsOfTrace")}
        </Link>
      </div>

      {loading ? (
        <Skeleton className="h-48 w-full" />
      ) : !detail || detail.spans.length === 0 ? (
        <p className="text-muted-foreground text-sm">{t("traces.noSpans")}</p>
      ) : (
        <>
          <div className="space-y-1">
            <h2 className="font-mono text-sm break-all">{detail.trace_id}</h2>
            <p className="text-muted-foreground text-sm">
              {t("traces.totalSpans", {
                duration: formatDuration(detail.duration_ms),
                count: detail.spans.length,
                source: sourceLabel(detail.source),
              })}
            </p>
          </div>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">{t("traces.waterfall")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {detail.spans.map((sp) => {
                const startMs = new Date(sp.start_ts).getTime() - origin
                const widthPct = Math.max((sp.duration_ms / totalWidth) * 100, sp.duration_ms > 0 ? 2 : 0.5)
                const leftPct = (startMs / totalWidth) * 100
                const barColor =
                  sp.status === "error" ? "bg-destructive/80" : sp.kind === "log" ? "bg-muted-foreground/50" : "bg-primary/70"
                return (
                  <div key={`${sp.span_id}-${sp.start_ts}`} className="grid gap-1 md:grid-cols-[200px_1fr_160px] md:items-center">
                    <div className="truncate text-xs">
                      <div className="font-medium">{sp.service || sp.container || "—"}</div>
                      <div className="text-muted-foreground truncate">{sp.name}</div>
                    </div>
                    <div className="bg-muted relative h-6 overflow-hidden rounded">
                      <div
                        className={`absolute top-0 h-full rounded ${barColor}`}
                        style={{ left: `${Math.max(leftPct, 0)}%`, width: `${Math.min(widthPct, 100 - leftPct)}%` }}
                        title={formatDuration(sp.duration_ms)}
                      />
                    </div>
                    <div className="text-muted-foreground flex flex-wrap items-center gap-2 text-xs">
                      <span className="tabular-nums">{formatDuration(sp.duration_ms)}</span>
                      {sp.container ? (
                        <Link className="text-primary hover:underline" to={`/logs?container=${encodeURIComponent(sp.container)}`}>
                          {t("traces.containerLogs")}
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
              <CardTitle className="text-sm font-medium">{t("traces.timeline")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2 text-xs">
              {detail.spans.map((sp) => (
                <div key={`tl-${sp.span_id}-${sp.start_ts}`} className="flex flex-wrap items-center gap-2 border-b pb-2 last:border-0">
                  <span className="font-mono tabular-nums">{formatClock(sp.start_ts, i18n.language)}</span>
                  <Badge variant="secondary">{sp.kind}</Badge>
                  <span className="font-medium">{sp.service}</span>
                  <span className="text-muted-foreground">{sp.name}</span>
                </div>
              ))}
            </CardContent>
          </Card>
        </>
      )}
    </section>
  )
}
