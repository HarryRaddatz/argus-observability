import { useCallback, useEffect, useState } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"

import { GroupSelect } from "@/components/filters/group-select"
import { TimeRangePicker } from "@/components/filters/time-range-picker"
import { PageHeader } from "@/components/layout/page-header"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { usePolling } from "@/hooks/use-polling"
import { useQueryPatch, useQueryState } from "@/hooks/use-query-state"
import { listWorkloads, searchLogs, type LogRow } from "@/lib/api"
import { LOG_LEVELS, LOG_TOPICS } from "@/lib/observability"
import { LogPatterns } from "@/views/log-patterns"

const levelVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  error: "destructive",
  warn: "secondary",
  warning: "secondary",
  debug: "outline",
  info: "outline",
}

function getTraceId(row: LogRow): string | undefined {
  const tid = row.fields?.trace_id
  return typeof tid === "string" && tid.length > 0 ? tid : undefined
}

function formatTraceShort(id: string): string {
  if (id.length <= 12) return id
  return `${id.slice(0, 8)}…`
}

export function LogsPage() {
  const { t } = useTranslation()
  const [mode] = useQueryState("mode", "lines")
  const [q] = useQueryState("q", "")
  const [traceId] = useQueryState("trace_id", "")
  const [since, setSince] = useQueryState("since", traceId ? "24h" : "1h")
  const [level, setLevel] = useQueryState("level", "all")
  const [topic, setTopic] = useQueryState("topic", "all")
  const [container, setContainer] = useQueryState("container", "all")
  const [group] = useQueryState("group", "")
  const patch = useQueryPatch()

  const [queryDraft, setQueryDraft] = useState(q)
  const [traceDraft, setTraceDraft] = useState(traceId)
  const [containers, setContainers] = useState<string[]>([])

  useEffect(() => setQueryDraft(q), [q])
  useEffect(() => setTraceDraft(traceId), [traceId])

  useEffect(() => {
    listWorkloads("1h")
      .then((wl) => setContainers(wl.map((w) => w.container).sort()))
      .catch(() => setContainers([]))
  }, [])

  const patterns = mode === "patterns"
  const traceActive = !patterns && traceId.length > 0

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("logs.title")}
        actions={
          <>
            <GroupSelect
              value={group}
              disabled={traceActive}
              onChange={(v) => patch({ group: v, container: v ? null : container })}
            />
            <select
              aria-label={t("common.container")}
              className="border-input bg-background h-8 rounded-md border px-2 text-sm disabled:opacity-50"
              value={container}
              disabled={Boolean(group) || traceActive}
              onChange={(e) => setContainer(e.target.value)}
            >
              <option value="all">{t("common.allContainers")}</option>
              {containers.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>
            <TimeRangePicker value={since} onChange={setSince} />
          </>
        }
      />

      <Tabs value={mode} onValueChange={(v) => patch({ mode: String(v) }, { mode: "lines" })}>
        <TabsList aria-label={t("common.mode")}>
          <TabsTrigger value="lines">{t("logs.lines")}</TabsTrigger>
          <TabsTrigger value="patterns">{t("logs.patterns")}</TabsTrigger>
        </TabsList>
      </Tabs>

      <div className="flex flex-wrap gap-3">
        <form
          role="search"
          className="flex min-w-[240px] flex-1 gap-2"
          onSubmit={(e) => {
            e.preventDefault()
            patch({ q: queryDraft.trim() })
          }}
        >
          <Input aria-label={t("logs.message")} placeholder={t("logs.message")} value={queryDraft} onChange={(e) => setQueryDraft(e.target.value)} />
          <Button type="submit" size="sm" variant="outline">
            {t("logs.search")}
          </Button>
        </form>
        <form
          className="flex min-w-[240px] flex-1 gap-2"
          onSubmit={(e) => {
            e.preventDefault()
            const tid = traceDraft.trim()
            patch({ trace_id: tid, since: tid && since === "1h" ? "24h" : since }, { since: "1h" })
          }}
        >
          <Input
            aria-label="Trace ID"
            placeholder="Trace ID"
            value={traceDraft}
            disabled={patterns}
            onChange={(e) => setTraceDraft(e.target.value)}
            className="font-mono text-xs"
          />
          <Button type="submit" size="sm" variant="outline" disabled={patterns}>
            {t("logs.filterTrace")}
          </Button>
        </form>
      </div>

      <div className="flex flex-wrap items-start gap-x-6 gap-y-3">
        <div role="group" aria-label={t("logs.topic")} className="space-y-1">
          <p className="text-muted-foreground text-sm">{t("logs.topic")}</p>
          <div className="flex flex-wrap gap-1">
            {LOG_TOPICS.map((item) => (
              <Button
                key={item.id}
                size="sm"
                variant={topic === item.id ? "default" : "outline"}
                aria-pressed={topic === item.id}
                disabled={patterns}
                onClick={() => setTopic(item.id)}
              >
                {item.id === "error" ? t("logs.errors") : t(`logs.${item.id}`)}
              </Button>
            ))}
          </div>
        </div>
        <div role="group" aria-label={t("logs.level")} className="space-y-1">
          <p className="text-muted-foreground text-sm">{t("logs.level")}</p>
          <div className="flex flex-wrap gap-1">
            {LOG_LEVELS.map((item) => (
              <Button
                key={item.id}
                size="sm"
                variant={level === item.id ? "secondary" : "ghost"}
                aria-pressed={level === item.id}
                disabled={patterns}
                onClick={() => setLevel(item.id)}
              >
                {item.id === "all" ? t("logs.all") : item.id}
              </Button>
            ))}
          </div>
        </div>
      </div>
      {patterns ? (
        <p className="text-muted-foreground text-xs">{t("logs.linesOnlyFilters")}</p>
      ) : null}

      {traceActive ? (
        <div className="bg-muted/50 flex flex-wrap items-center gap-2 rounded-md border px-3 py-2 text-sm">
          <span>
            Linhas do trace <span className="font-mono">{traceId}</span> em todos os containers.
          </span>
          <Link to={`/traces?trace_id=${encodeURIComponent(traceId)}`} className="text-primary hover:underline">
            Ver o trace
          </Link>
          <Button size="sm" variant="ghost" onClick={() => patch({ trace_id: null })}>
            Limpar trace
          </Button>
        </div>
      ) : null}

      {patterns ? (
        <LogPatterns
          filters={{ since, q, container, group }}
          onOpenLines={(p) => patch({ mode: null, q: p.pattern.slice(0, 40), container: p.container, group: null })}
        />
      ) : (
        <LogLines
          q={q}
          since={since}
          level={level}
          topic={topic}
          container={container}
          group={group}
          traceId={traceId}
        />
      )}
    </div>
  )
}

type LinesProps = {
  q: string
  since: string
  level: string
  topic: string
  container: string
  group: string
  traceId: string
}

function LogLines({ q, since, level, topic, container, group, traceId }: LinesProps) {
  const { t } = useTranslation()
  const [rows, setRows] = useState<LogRow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    searchLogs({
      q,
      since,
      level,
      topic,
      trace_id: traceId || undefined,
      container: traceId || group ? undefined : container,
      group: traceId ? undefined : group || undefined,
    })
      .then((r) => {
        setRows(r)
        setError(null)
      })
      .catch((e) => {
        setError(e instanceof Error ? e.message : t("logs.searchError"))
        setRows([])
      })
      .finally(() => setLoading(false))
  }, [q, since, level, topic, container, group, traceId, t])

  usePolling(load, 15_000)

  return (
    <>
      {error ? <p className="text-destructive text-sm">{error}</p> : null}
      <ScrollArea className="h-[480px] rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("logs.time")}</TableHead>
              <TableHead>{t("logs.level")}</TableHead>
              <TableHead>{t("common.container")}</TableHead>
              <TableHead>{t("logs.trace")}</TableHead>
              <TableHead>{t("logs.topics")}</TableHead>
              <TableHead>{t("logs.message")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6}>
                  <Skeleton className="h-8 w-full" />
                </TableCell>
              </TableRow>
            ) : rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-muted-foreground text-center">
                  {t("logs.linesEmpty")}
                </TableCell>
              </TableRow>
            ) : (
              rows.map((row, i) => <LogRowView key={`${row.ts}-${row.entity_uid}-${i}`} row={row} />)
            )}
          </TableBody>
        </Table>
      </ScrollArea>
    </>
  )
}

function LogRowView({ row }: { row: LogRow }) {
  const { t, i18n } = useTranslation()
  const topics = (row.fields?.topics as string[] | undefined) ?? []
  const tid = getTraceId(row)
  const container = row.entity_uid.split(":").pop() ?? row.entity_uid
  return (
    <TableRow>
      <TableCell className="whitespace-nowrap font-mono text-xs">{new Date(row.ts).toLocaleTimeString(i18n.language)}</TableCell>
      <TableCell>
        <Badge variant={levelVariant[row.level] ?? "outline"}>{row.level}</Badge>
      </TableCell>
      <TableCell className="max-w-[160px] truncate font-mono text-xs">
        <Link to={`/metrics?container=${encodeURIComponent(container)}`} className="hover:underline" title={t("logs.metricsOf", { container })}>
          {container}
        </Link>
      </TableCell>
      <TableCell className="font-mono text-xs">
        {tid ? (
          <Link to={`/traces?trace_id=${encodeURIComponent(tid)}`} className="text-primary hover:underline" title={tid}>
            {formatTraceShort(tid)}
          </Link>
        ) : (
          <span className="text-muted-foreground">—</span>
        )}
      </TableCell>
      <TableCell>
        <div className="flex flex-wrap gap-1">
          {topics
            .filter((topicId) => topicId !== "general")
            .map((topicId) => (
              <Badge key={topicId} variant="outline" className="text-xs">
                {topicId}
              </Badge>
            ))}
        </div>
      </TableCell>
      <TableCell className="max-w-[520px] font-mono text-xs break-all whitespace-pre-wrap">{row.message}</TableCell>
    </TableRow>
  )
}
