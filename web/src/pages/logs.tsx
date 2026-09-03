import { useCallback, useEffect, useState } from "react"
import { useSearchParams } from "react-router-dom"

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
import { listWorkloads, searchLogs, type LogRow } from "@/lib/api"
import { LOG_LEVELS, LOG_TOPICS, TIME_RANGES } from "@/lib/observability"

const levelVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  error: "destructive",
  warn: "secondary",
  warning: "secondary",
  debug: "outline",
  info: "outline",
}

export function LogsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [query, setQuery] = useState(searchParams.get("q") ?? "")
  const [level, setLevel] = useState(searchParams.get("level") ?? "all")
  const [topic, setTopic] = useState(searchParams.get("topic") ?? "all")
  const [container, setContainer] = useState(searchParams.get("container") ?? "all")
  const [since, setSince] = useState(searchParams.get("since") ?? "1h")
  const [containers, setContainers] = useState<string[]>([])
  const [rows, setRows] = useState<LogRow[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    listWorkloads("1h")
      .then((wl) => setContainers(wl.map((w) => w.container).sort()))
      .catch(() => setContainers([]))
  }, [])

  const runSearch = useCallback(() => {
    setLoading(true)
    setError(null)
    searchLogs({ q: query, since, level, container, topic })
      .then(setRows)
      .catch((e) => {
        setError(e instanceof Error ? e.message : "Erro")
        setRows([])
      })
      .finally(() => setLoading(false))
  }, [query, since, level, container, topic])

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
    if (container !== "all") p.set("container", container)
    if (since !== "1h") p.set("since", since)
    setSearchParams(p, { replace: true })
  }, [query, level, topic, container, since, setSearchParams])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Logs</h1>
        <p className="text-muted-foreground text-sm">
          Filtros inteligentes por tópico (GC, memória, OOM), level e container.
        </p>
      </div>

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
        <select
          className="border-input bg-background h-9 rounded-md border px-3 text-sm"
          value={container}
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
              <TableHead>Tópicos</TableHead>
              <TableHead>Mensagem</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5}>
                  <Skeleton className="h-8 w-full" />
                </TableCell>
              </TableRow>
            ) : rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="text-muted-foreground text-center">
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
