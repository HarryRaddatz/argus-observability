import { useEffect, useState } from "react"
import { Link } from "react-router-dom"

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
import { listWorkloads, type WorkloadSnapshot } from "@/lib/api"
import { formatBytes, formatPercent } from "@/lib/format"

export function WorkloadsPage() {
  const [rows, setRows] = useState<WorkloadSnapshot[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    function load() {
      listWorkloads("30m")
        .then((data) => {
          if (!cancelled) {
            setRows([...data].sort((a, b) => a.container.localeCompare(b.container)))
            setError(null)
          }
        })
        .catch((e) => {
          if (!cancelled) setError(e instanceof Error ? e.message : "Erro")
        })
        .finally(() => {
          if (!cancelled) setLoading(false)
        })
    }
    load()
    const t = setInterval(load, 30_000)
    return () => {
      cancelled = true
      clearInterval(t)
    }
  }, [])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Workloads</h1>
        <p className="text-muted-foreground text-sm">Containers monitorados pelo agent.</p>
      </div>

      {error ? <p className="text-destructive text-sm">{error}</p> : null}

      <ScrollArea className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Container</TableHead>
              <TableHead>CPU</TableHead>
              <TableHead>Memória</TableHead>
              <TableHead>Atualizado</TableHead>
              <TableHead />
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
                  Nenhum container monitorado.
                </TableCell>
              </TableRow>
            ) : (
              rows.map((w) => (
                <TableRow key={w.entity_uid}>
                  <TableCell className="font-medium">{w.container}</TableCell>
                  <TableCell>
                    <Badge variant={w.cpu_usage > 80 ? "destructive" : "secondary"}>
                      {formatPercent(w.cpu_usage)}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-sm">
                    {formatBytes(w.memory_usage)}
                    {w.memory_limit > 0 ? (
                      <span className="text-muted-foreground"> / {formatBytes(w.memory_limit)}</span>
                    ) : null}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-xs whitespace-nowrap">
                    {new Date(w.updated_at).toLocaleString("pt-BR")}
                  </TableCell>
                  <TableCell>
                    <Link
                      to={`/metrics?container=${encodeURIComponent(w.container)}`}
                      className="text-primary text-sm hover:underline"
                    >
                      Gráficos
                    </Link>
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
