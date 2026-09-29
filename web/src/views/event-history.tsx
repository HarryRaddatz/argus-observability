import { useCallback, useState } from "react"
import { Link } from "react-router-dom"

import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { usePolling } from "@/hooks/use-polling"
import { listEvents, type EventRow } from "@/lib/api"
import { severityLabel, severityVariant } from "@/lib/severity"

function formatPayload(payload?: Record<string, unknown>) {
  if (!payload || Object.keys(payload).length === 0) return "—"
  return Object.entries(payload)
    .slice(0, 3)
    .map(([k, v]) => `${k}=${String(v)}`)
    .join(", ")
}

export function EventHistory({ since }: { since: string }) {
  const [rows, setRows] = useState<EventRow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    listEvents(undefined, since)
      .then((ev) => {
        setRows(ev)
        setError(null)
      })
      .catch((e) => setError(e instanceof Error ? e.message : "Não foi possível carregar os eventos"))
      .finally(() => setLoading(false))
  }, [since])

  usePolling(load)

  return (
    <>
      {error ? <p className="text-destructive text-sm">{error}</p> : null}
      <ScrollArea className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Hora</TableHead>
              <TableHead>Tipo</TableHead>
              <TableHead>Severidade</TableHead>
              <TableHead>Container</TableHead>
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
                  Nenhum evento no período. Amplie o período para ver mais.
                </TableCell>
              </TableRow>
            ) : (
              rows.map((e) => {
                const container = e.entity_uid.split(":").pop() ?? e.entity_uid
                return (
                  <TableRow key={e.id}>
                    <TableCell className="whitespace-nowrap font-mono text-xs">{new Date(e.ts).toLocaleString("pt-BR")}</TableCell>
                    <TableCell className="font-medium">{e.type}</TableCell>
                    <TableCell>
                      <Badge variant={severityVariant[e.severity] ?? "outline"}>{severityLabel[e.severity] ?? e.severity}</Badge>
                    </TableCell>
                    <TableCell className="max-w-[180px] truncate font-mono text-xs">
                      <Link to={`/logs?container=${encodeURIComponent(container)}`} className="text-primary hover:underline">
                        {container}
                      </Link>
                    </TableCell>
                    <TableCell className="text-muted-foreground text-sm">{e.source}</TableCell>
                    <TableCell className="text-muted-foreground max-w-[240px] truncate text-xs">{formatPayload(e.payload)}</TableCell>
                  </TableRow>
                )
              })
            )}
          </TableBody>
        </Table>
      </ScrollArea>
    </>
  )
}
