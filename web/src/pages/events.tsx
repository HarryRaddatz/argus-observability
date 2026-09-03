import { useCallback, useEffect, useState } from "react"

import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
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
import { fetchActiveAlerts, listEvents, type ActiveAlert, type EventRow } from "@/lib/api"

const severityVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  critical: "destructive",
  warning: "secondary",
  info: "outline",
  debug: "outline",
}

function formatPayload(payload?: Record<string, unknown>) {
  if (!payload || Object.keys(payload).length === 0) return "—"
  return Object.entries(payload)
    .slice(0, 3)
    .map(([k, v]) => `${k}=${String(v)}`)
    .join(", ")
}

export function EventsPage() {
  const [rows, setRows] = useState<EventRow[]>([])
  const [active, setActive] = useState<ActiveAlert[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    Promise.all([listEvents(undefined, "24h"), fetchActiveAlerts()])
      .then(([ev, alerts]) => {
        setRows(ev)
        setActive(alerts)
        setError(null)
      })
      .catch((e) => setError(e instanceof Error ? e.message : "Erro"))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    load()
    const t = setInterval(load, 15_000)
    return () => clearInterval(t)
  }, [load])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Eventos</h1>
        <p className="text-muted-foreground text-sm">
          Lifecycle, rule engine e alertas ativos — atualiza a cada 15s.
        </p>
      </div>

      {active.length > 0 ? (
        <div className="space-y-2">
          <h2 className="text-sm font-medium">Alertas ativos</h2>
          <div className="grid gap-3 md:grid-cols-2">
            {active.map((a) => (
              <Card key={a.rule_id + a.entity_uid} className="border-destructive/30">
                <CardHeader className="pb-2">
                  <div className="flex items-center justify-between gap-2">
                    <CardTitle className="text-sm">{a.title}</CardTitle>
                    <Badge variant={severityVariant[a.severity] ?? "outline"}>ativo</Badge>
                  </div>
                </CardHeader>
                <CardContent className="text-muted-foreground text-xs">{a.summary}</CardContent>
              </Card>
            ))}
          </div>
        </div>
      ) : null}

      {error ? <p className="text-destructive text-sm">{error}</p> : null}
      <ScrollArea className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Hora</TableHead>
              <TableHead>Tipo</TableHead>
              <TableHead>Severidade</TableHead>
              <TableHead>Entidade</TableHead>
              <TableHead>Origem</TableHead>
              <TableHead>Detalhe</TableHead>
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
                  Nenhum evento no período.
                </TableCell>
              </TableRow>
            ) : (
              rows.map((e) => (
                <TableRow key={e.id}>
                  <TableCell className="whitespace-nowrap font-mono text-xs">
                    {new Date(e.ts).toLocaleString("pt-BR")}
                  </TableCell>
                  <TableCell className="font-medium">{e.type}</TableCell>
                  <TableCell>
                    <Badge variant={severityVariant[e.severity] ?? "outline"}>{e.severity}</Badge>
                  </TableCell>
                  <TableCell className="max-w-[180px] truncate font-mono text-xs">
                    {e.entity_uid.split(":").pop()}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-sm">{e.source}</TableCell>
                  <TableCell className="text-muted-foreground max-w-[240px] truncate text-xs">
                    {formatPayload(e.payload)}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </ScrollArea>
    </div>
  )
}
