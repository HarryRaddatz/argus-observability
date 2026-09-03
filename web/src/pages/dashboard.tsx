import { useEffect, useState } from "react"

import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { getHealth, listEvents } from "@/lib/api"

export function DashboardPage() {
  const [health, setHealth] = useState<"ok" | "error" | "loading">("loading")
  const [events, setEvents] = useState<number | null>(null)

  useEffect(() => {
    getHealth()
      .then(() => setHealth("ok"))
      .catch(() => setHealth("error"))
    listEvents(undefined, "24h")
      .then((rows) => setEvents(rows.length))
      .catch(() => setEvents(0))
  }, [])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Visão geral</h1>
        <p className="text-muted-foreground text-sm">
          Estado do hub, agents e eventos recentes.
        </p>
      </div>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>Hub</CardTitle>
            <CardDescription>Disponibilidade da API</CardDescription>
          </CardHeader>
          <CardContent>
            {health === "loading" ? (
              <Skeleton className="h-6 w-20" />
            ) : (
              <Badge variant={health === "ok" ? "default" : "destructive"}>
                {health === "ok" ? "Online" : "Indisponível"}
              </Badge>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Eventos</CardTitle>
            <CardDescription>Últimas 24 horas</CardDescription>
          </CardHeader>
          <CardContent>
            {events === null ? (
              <Skeleton className="h-8 w-12" />
            ) : (
              <p className="text-3xl font-semibold tabular-nums">{events}</p>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Agents</CardTitle>
            <CardDescription>Conectados ao hub</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground text-sm">
              Detalhe na view Workloads quando a API expuser a lista.
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
