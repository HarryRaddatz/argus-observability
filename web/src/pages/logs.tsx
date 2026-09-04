import { useCallback, useEffect, useState } from "react"
import { Link, useSearchParams } from "react-router-dom"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { listWorkloads, listWorkloadGroups, searchLogs, type LogRow, type WorkloadGroup } from "@/lib/api"
import { LOG_LEVELS, LOG_TOPICS, TIME_RANGES } from "@/lib/observability"

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
  const [searchParams, setSearchParams] = useSearchParams()
  const [query, setQuery] = useState(searchParams.get("q") ?? "")
  const [level, setLevel] = useState(searchParams.get("level") ?? "all")
  const [topic, setTopic] = useState(searchParams.get("topic") ?? "all")
  const [container, setContainer] = useState(searchParams.get("container") ?? "all")
  const [group, setGroup] = useState(searchParams.get("group") ?? "")
  const [traceId, setTraceId] = useState(searchParams.get("trace_id") ?? "")
  const [since, setSince] = useState(() => {
    const s = searchParams.get("since")
    if (s) return s
    if (searchParams.get("trace_id")) return "24h"
    return "1h"
  })
  const [containers, setContainers] = useState<string[]>([])
  const [groups, setGroups] = useState<WorkloadGroup[]>([])
  const [rows, setRows] = useState<LogRow[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const traceFilterActive = traceId.trim().length > 0

  useEffect(() => {
    Promise.all([listWorkloads("1h"), listWorkloadGroups()])
      .then(([wl, g]) => {
        setContainers(wl.map((w) => w.container).sort())
        setGroups(g)
      })
      .catch(() => {
        setContainers([])
        setGroups([])
      })
  }, [])

  const runSearch = useCallback(() => {
    setLoading(true)
    setError(null)
    const tid = traceId.trim()
    searchLogs({
      q: query,
      since,
      level,
      topic,
      trace_id: tid || undefined,
      container: tid || group ? undefined : container,
      group: tid ? undefined : group || undefined,
    })
      .then(setRows)
      .catch((e) => {
        setError(e instanceof Error ? e.message : "Erro")
        setRows([])
      })
      .finally(() => setLoading(false))
  }, [query, since, level, container, topic, group, traceId])

  useEffect(() => {
    runSearch()
    const t = setInterval(runSearch, 15_000)
    return () => clearInterval(t)
  }, [runSearch])

  useEffect(() => {
    const p = new URLSearchParams()
    if (query) p.set("q", query)
    if (level !== "all") p.set("level", level)
    if (topic !== "all") p.set("topic", topic)
    if (container !== "all" && !traceFilterActive) p.set("container", container)
    if (group && !traceFilterActive) p.set("group", group)
    if (traceFilterActive) p.set("trace_id", traceId.trim())
    if (since !== "1h") p.set("since", since)
    setSearchParams(p, { replace: true })
  }, [query, level, topic, container, group, traceId, traceFilterActive, since, setSearchParams])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Logs</h1>
        <p className="text-muted-foreground text-sm">
          Filtros por tópico, level, container e traceId cross-container.
        </p>
      </div>

      {traceFilterActive ? (
        <div className="bg-muted/50 flex flex-wrap items-center gap-2 rounded-md border px-3 py-2 text-sm">
          <span>
            Trace <span className="font-mono">{traceId.trim()}</span> — todos os containers
          </span>
          <Button
            size="sm"
            variant="ghost"
            onClick={() => {
              setTraceId("")
            }}
          >
            Limpar trace
          </Button>
        </div>
      ) : null}

      <div className="flex flex-wrap gap-4">
        <form
          className="min-w-[200px] flex-1"
          onSubmit={(e) => {
            e.preventDefault()
            runSearch()
          }}
        >
          <Input
            placeholder="Texto livre na mensagem..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </form>
        <form
          className="min-w-[220px] flex-1"
          onSubmit={(e) => {
            e.preventDefault()
            if (traceId.trim() && since === "1h") setSince("24h")
            runSearch()
          }}
        >
          <Input
            placeholder="Trace ID (UUID ou hex)..."
            value={traceId}
            onChange={(e) => setTraceId(e.target.value)}
            className="font-mono text-xs"
          />
        </form>
        <select
          className="border-input bg-background h-9 rounded-md border px-3 text-sm"
          value={group}
          disabled={traceFilterActive}
          onChange={(e) => {
            setGroup(e.target.value)
            if (e.target.value) setContainer("all")
          }}
        >
          <option value="">Todos grupos</option>
          {groups.map((g) => (
            <option key={g.id} value={g.id}>
              {g.name}
            </option>
          ))}
        </select>
        <select
          className="border-input bg-background h-9 rounded-md border px-3 text-sm disabled:opacity-50"
          value={container}
          disabled={Boolean(group) || traceFilterActive}
          onChange={(e) => setContainer(e.target.value)}
        >
          <option value="all">Todos containers</option>
          {containers.map((c) => (
            <option key={c} value={c}>
              {c}
            </option>
          ))}
        </select>
      </div>

      <div className="space-y-2">
        <p className="text-muted-foreground text-xs font-medium uppercase">Tópico</p>
        <div className="flex flex-wrap gap-1">
          {LOG_TOPICS.map((t) => (
            <Button
              key={t.id}
              size="sm"
              variant={topic === t.id ? "default" : "outline"}
              onClick={() => setTopic(t.id)}
            >
              {t.label}
            </Button>
          ))}
        </div>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-4">
        <div className="flex flex-wrap gap-1">
          {LOG_LEVELS.map((l) => (
            <Button
              key={l.id}
              size="sm"
              variant={level === l.id ? "secondary" : "ghost"}
              onClick={() => setLevel(l.id)}
            >
              {l.label}
            </Button>
          ))}
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

      {error ? <p className="text-destructive text-sm">{error}</p> : null}

      <ScrollArea className="h-[480px] rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Hora</TableHead>
              <TableHead>Level</TableHead>
              <TableHead>Container</TableHead>
              <TableHead>Trace</TableHead>
              <TableHead>Tópicos</TableHead>
              <TableHead>Mensagem</TableHead>
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
                  Nenhuma linha com os filtros atuais.
                </TableCell>
              </TableRow>
            ) : (
              rows.map((row, i) => (
                <LogRowView key={`${row.ts}-${row.entity_uid}-${i}`} row={row} />
              ))
            )}
          </TableBody>
        </Table>
      </ScrollArea>
    </div>
  )
}

function LogRowView({ row }: { row: LogRow }) {
  const topics = (row.fields?.topics as string[] | undefined) ?? []
  const tid = getTraceId(row)
  return (
    <TableRow>
      <TableCell className="whitespace-nowrap font-mono text-xs">
        {new Date(row.ts).toLocaleTimeString("pt-BR")}
      </TableCell>
      <TableCell>
        <Badge variant={levelVariant[row.level] ?? "outline"}>{row.level}</Badge>
      </TableCell>
      <TableCell className="max-w-[140px] truncate font-mono text-xs">
        {row.entity_uid.split(":").pop()}
      </TableCell>
      <TableCell className="font-mono text-xs">
        {tid ? (
          <Link
            to={`/traces?trace_id=${encodeURIComponent(tid)}`}
            className="text-primary hover:underline"
            title={tid}
          >
            {formatTraceShort(tid)}
          </Link>
        ) : (
          <span className="text-muted-foreground">—</span>
        )}
      </TableCell>
      <TableCell>
        <div className="flex flex-wrap gap-1">
          {topics.filter((t) => t !== "general").map((t) => (
            <Badge key={t} variant="outline" className="text-[10px]">
              {t}
            </Badge>
          ))}
        </div>
      </TableCell>
      <TableCell className="max-w-[520px] font-mono text-xs whitespace-pre-wrap break-all">{row.message}</TableCell>
    </TableRow>
  )
}
