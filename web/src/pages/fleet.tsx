import { useEffect, useState } from "react"
import { Link } from "react-router-dom"

import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
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
import { cn } from "@/lib/utils"
import { fetchFleetStatus, type ContainerFleetStatus, type FleetStatusResponse } from "@/lib/api"

const stateVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  running: "default",
  restarting: "secondary",
  exited: "outline",
  dead: "destructive",
}

export function FleetPage() {
  const [data, setData] = useState<FleetStatusResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    function load() {
      fetchFleetStatus()
        .then((resp) => {
          if (!cancelled) {
            setData(resp)
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

  const summary = data?.summary
  const services = data?.services ?? []
  const containers = data?.containers ?? []

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Fleet</h1>
        <p className="text-muted-foreground text-sm">
          Estado operacional — restarts, OOM, healthcheck e réplicas por serviço.
        </p>
      </div>

      {error ? <p className="text-destructive text-sm">{error}</p> : null}

      {loading && !data ? (
        <Skeleton className="h-24 w-full" />
      ) : summary ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
          <StatCard label="Running" value={summary.running} />
          <StatCard label="Restarting" value={summary.restarting} warn={summary.restarting > 0} />
          <StatCard label="Unhealthy" value={summary.unhealthy} warn={summary.unhealthy > 0} />
          <StatCard label="Exited / Dead" value={summary.exited + summary.dead} />
          <StatCard label="Restarts (total)" value={summary.total_restart_count} warn={summary.total_restart_count > 10} />
        </div>
      ) : null}

      {data?.events_24h ? (
        <p className="text-muted-foreground text-xs">
          Eventos 24h: {data.events_24h.restarts_24h} restarts · {data.events_24h.oom_24h} OOM ·{" "}
          {data.events_24h.failures_24h} falhas
        </p>
      ) : null}

      {services.length > 0 ? (
        <section className="space-y-2">
          <h2 className="text-sm font-medium">Serviços</h2>
          <ScrollArea className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Serviço</TableHead>
                  <TableHead>Réplicas</TableHead>
                  <TableHead>Restarting</TableHead>
                  <TableHead>Unhealthy</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {services.map((s) => (
                  <TableRow key={s.service}>
                    <TableCell className="font-medium">{s.service}</TableCell>
                    <TableCell>
                      {s.replicas_up}/{s.replicas_total}
                    </TableCell>
                    <TableCell>{s.restarting || "—"}</TableCell>
                    <TableCell>{s.unhealthy || "—"}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </ScrollArea>
        </section>
      ) : null}

      <section className="space-y-2">
        <h2 className="text-sm font-medium">Containers</h2>
        <ScrollArea className="h-[min(480px,60vh)] rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Container</TableHead>
                <TableHead>Estado</TableHead>
                <TableHead>Health</TableHead>
                <TableHead>Restarts</TableHead>
                <TableHead>OOM</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading && containers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6}>
                    <Skeleton className="h-8 w-full" />
                  </TableCell>
                </TableRow>
              ) : containers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-muted-foreground text-center">
                    Aguardando primeiro snapshot do agent.
                  </TableCell>
                </TableRow>
              ) : (
                containers.map((c) => <FleetRow key={c.entity_uid} row={c} />)
              )}
            </TableBody>
          </Table>
        </ScrollArea>
      </section>
    </div>
  )
}

function StatCard({ label, value, warn }: { label: string; value: number; warn?: boolean }) {
  return (
    <Card>
      <CardContent className="pt-6">
        <p className="text-muted-foreground text-xs">{label}</p>
        <p className={cn("text-2xl font-semibold tabular-nums", warn && "text-amber-600")}>{value}</p>
      </CardContent>
    </Card>
  )
}

function FleetRow({ row }: { row: ContainerFleetStatus }) {
  const state = row.state?.toLowerCase() ?? "unknown"
  return (
    <TableRow className={row.oom_killed ? "bg-destructive/5" : undefined}>
      <TableCell className="font-medium">{row.container}</TableCell>
      <TableCell>
        <Badge variant={stateVariant[state] ?? "outline"}>{row.state}</Badge>
      </TableCell>
      <TableCell className="text-muted-foreground text-xs">{row.health || "—"}</TableCell>
      <TableCell className={row.restart_count > 3 ? "text-amber-600 font-medium" : ""}>
        {row.restart_count}
      </TableCell>
      <TableCell>{row.oom_killed ? <Badge variant="destructive">OOM</Badge> : "—"}</TableCell>
      <TableCell>
        <Link
          to={`/logs?container=${encodeURIComponent(row.container)}`}
          className="text-primary text-xs hover:underline"
        >
          Logs
        </Link>
      </TableCell>
    </TableRow>
  )
}
