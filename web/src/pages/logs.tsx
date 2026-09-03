import { useEffect, useState } from "react"

import { Badge } from "@/components/ui/badge"
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
import { searchLogs, type LogRow } from "@/lib/api"

export function LogsPage() {
  const [query, setQuery] = useState("")
  const [rows, setRows] = useState<LogRow[]>([])
  const [loading, setLoading] = useState(false)

  function runSearch(q: string) {
    setLoading(true)
    searchLogs(q, "15m")
      .then(setRows)
      .catch(() => setRows([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    runSearch("")
  }, [])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Logs</h1>
        <p className="text-muted-foreground text-sm">
          Busca nas entradas indexadas pelo hub.
        </p>
      </div>
      <form
        className="flex max-w-md gap-2"
        onSubmit={(e) => {
          e.preventDefault()
          runSearch(query)
        }}
      >
        <Input
          placeholder="Filtrar mensagem..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
      </form>
      <ScrollArea className="h-[480px] rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Hora</TableHead>
              <TableHead>Level</TableHead>
              <TableHead>Entidade</TableHead>
              <TableHead>Mensagem</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={4}>
                  <Skeleton className="h-8 w-full" />
                </TableCell>
              </TableRow>
            ) : rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className="text-muted-foreground text-center">
                  Nenhuma linha encontrada.
                </TableCell>
              </TableRow>
            ) : (
              rows.map((row, i) => (
                <TableRow key={`${row.ts}-${i}`}>
                  <TableCell className="whitespace-nowrap font-mono text-xs">
                    {new Date(row.ts).toLocaleTimeString()}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">{row.level}</Badge>
                  </TableCell>
                  <TableCell className="max-w-[160px] truncate font-mono text-xs">
                    {row.entity_uid}
                  </TableCell>
                  <TableCell className="font-mono text-xs">{row.message}</TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </ScrollArea>
    </div>
  )
}
