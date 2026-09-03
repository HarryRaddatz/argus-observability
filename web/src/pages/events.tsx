import { useEffect, useState } from "react"

import { Badge } from "@/components/ui/badge"
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
import { listEvents, type EventRow } from "@/lib/api"

const severityVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  critical: "destructive",
  warning: "secondary",
  info: "outline",
  debug: "outline",
}

export function EventsPage() {
  const [rows, setRows] = useState<EventRow[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    listEvents(undefined, "24h")
      .then(setRows)
      .catch(() => setRows([]))
      .finally(() => setLoading(false))
  }, [])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Eventos</h1>
        <p className="text-muted-foreground text-sm">
          Timeline do barramento — alertas, desconexões e sinais de recurso.
        </p>
      </div>
      <ScrollArea className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Hora</TableHead>
              <TableHead>Tipo</TableHead>
              <TableHead>Severidade</TableHead>
              <TableHead>Entidade</TableHead>
              <TableHead>Origem</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={5}>
                  <Skeleton className="h-8 w-full" />
                </TableCell>
              </TableRow>
            ) : rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="text-muted-foreground text-center">
                  Nenhum evento no período.
                </TableCell>
              </TableRow>
            ) : (
              rows.map((e) => (
                <TableRow key={e.id}>
                  <TableCell className="whitespace-nowrap font-mono text-xs">
                    {new Date(e.ts).toLocaleString()}
                  </TableCell>
                  <TableCell className="font-medium">{e.type}</TableCell>
                  <TableCell>
                    <Badge variant={severityVariant[e.severity] ?? "outline"}>{e.severity}</Badge>
                  </TableCell>
                  <TableCell className="max-w-[200px] truncate font-mono text-xs">
                    {e.entity_uid}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-sm">{e.source}</TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </ScrollArea>
    </div>
  )
}
