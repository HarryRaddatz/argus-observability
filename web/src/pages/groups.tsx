import { useCallback, useEffect, useMemo, useState } from "react"
import { Link } from "react-router-dom"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
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
import { cn } from "@/lib/utils"
import {
  createWorkloadGroup,
  deleteWorkloadGroup,
  discoverWorkloadGroups,
  getWorkloadGroupSummary,
  listWorkloads,
  listWorkloadGroups,
  type WorkloadGroup,
  type WorkloadGroupInput,
  type WorkloadGroupSummary,
} from "@/lib/api"
import { formatBytes, formatPercent } from "@/lib/format"

const kindLabel: Record<string, string> = {
  stack: "Stack",
  service: "Serviço",
  custom: "Personalizado",
}

export function GroupsPage() {
  const [groups, setGroups] = useState<WorkloadGroup[]>([])
  const [discovered, setDiscovered] = useState<WorkloadGroup[]>([])
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [summary, setSummary] = useState<WorkloadGroupSummary | null>(null)
  const [containers, setContainers] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [customName, setCustomName] = useState("")
  const [customSelected, setCustomSelected] = useState<string[]>([])

  const load = useCallback(() => {
    setLoading(true)
    Promise.all([listWorkloadGroups(), discoverWorkloadGroups(), listWorkloads("1h")])
      .then(([g, d, wl]) => {
        setGroups(g)
        setDiscovered(d)
        setContainers(wl.map((w) => w.container).sort())
        setError(null)
      })
      .catch((e) => setError(e instanceof Error ? e.message : "Erro"))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    if (!selectedId) {
      setSummary(null)
      return
    }
    getWorkloadGroupSummary(selectedId, "30m")
      .then(setSummary)
      .catch(() => setSummary(null))
  }, [selectedId])

  const selectedGroup = useMemo(
    () => groups.find((g) => g.id === selectedId) ?? null,
    [groups, selectedId],
  )

  async function importDiscovered(g: WorkloadGroup) {
    const input: WorkloadGroupInput = {
      name: g.name,
      kind: g.kind as WorkloadGroupInput["kind"],
      label_key: g.label_key,
      label_value: g.label_value,
      description: `Importado automaticamente (${g.kind}: ${g.label_value})`,
    }
    await createWorkloadGroup(input)
    load()
  }

  async function saveCustomGroup() {
    if (!customName.trim() || customSelected.length === 0) return
    await createWorkloadGroup({
      name: customName.trim(),
      kind: "custom",
      containers: customSelected,
      description: "Grupo personalizado",
    })
    setCustomName("")
    setCustomSelected([])
    load()
  }

  async function removeGroup(id: string) {
    await deleteWorkloadGroup(id)
    if (selectedId === id) setSelectedId(null)
    load()
  }

  function toggleCustomContainer(name: string) {
    setCustomSelected((prev) =>
      prev.includes(name) ? prev.filter((c) => c !== name) : [...prev, name],
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Grupos de workload</h1>
        <p className="text-muted-foreground text-sm">
          Agrupe por stack, serviço Compose ou seleção manual para monitorar em conjunto.
        </p>
      </div>

      {error ? <p className="text-destructive text-sm">{error}</p> : null}

      <div className="grid gap-6 lg:grid-cols-[320px_1fr]">
        <aside className="space-y-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm">Seus grupos</CardTitle>
            </CardHeader>
            <CardContent className="space-y-1">
              {loading ? (
                <Skeleton className="h-20 w-full" />
              ) : groups.length === 0 ? (
                <p className="text-muted-foreground text-sm">Nenhum grupo salvo.</p>
              ) : (
                groups.map((g) => (
                  <button
                    key={g.id}
                    type="button"
                    className={cn(
                      "hover:bg-muted flex w-full flex-col rounded-md border px-3 py-2 text-left text-sm",
                      selectedId === g.id && "bg-muted border-primary/40",
                    )}
                    onClick={() => setSelectedId(g.id)}
                  >
                    <span className="font-medium">{g.name}</span>
                    <span className="text-muted-foreground text-xs">
                      {kindLabel[g.kind] ?? g.kind} · {g.member_count ?? 0} membros
                    </span>
                  </button>
                ))
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm">Sugeridos (stack/serviço)</CardTitle>
              <CardDescription>Detectados a partir dos labels atuais</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2">
              {discovered.slice(0, 8).map((g) => (
                <div key={g.id} className="flex items-center justify-between gap-2 text-sm">
                  <div className="min-w-0">
                    <p className="truncate font-medium">{g.name}</p>
                    <p className="text-muted-foreground text-xs">{g.member_count} containers</p>
                  </div>
                  <Button size="sm" variant="outline" onClick={() => importDiscovered(g)}>
                    Adicionar
                  </Button>
                </div>
              ))}
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm">Grupo personalizado</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <Input
                placeholder="Nome do grupo"
                value={customName}
                onChange={(e) => setCustomName(e.target.value)}
              />
              <ScrollArea className="h-36 rounded-md border">
                <ul className="p-1">
                  {containers.map((c) => (
                    <li key={c}>
                      <button
                        type="button"
                        className={cn(
                          "hover:bg-muted w-full rounded px-2 py-1.5 text-left text-xs",
                          customSelected.includes(c) && "bg-muted font-medium",
                        )}
                        onClick={() => toggleCustomContainer(c)}
                      >
                        {c}
                      </button>
                    </li>
                  ))}
                </ul>
              </ScrollArea>
              <Button size="sm" className="w-full" onClick={saveCustomGroup}>
                Criar ({customSelected.length} selecionados)
              </Button>
            </CardContent>
          </Card>
        </aside>

        <div className="space-y-4">
          {!selectedGroup ? (
            <Card>
              <CardContent className="text-muted-foreground pt-6 text-sm">
                Selecione um grupo para ver métricas agregadas e membros.
              </CardContent>
            </Card>
          ) : (
            <>
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <h2 className="text-xl font-semibold">{selectedGroup.name}</h2>
                  <div className="mt-1 flex gap-2">
                    <Badge variant="outline">{kindLabel[selectedGroup.kind]}</Badge>
                    {selectedGroup.label_value ? (
                      <Badge variant="secondary">{selectedGroup.label_value}</Badge>
                    ) : null}
                  </div>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Link
                    to={`/explorer?group=${encodeURIComponent(selectedGroup.id)}`}
                    className="border-input bg-background hover:bg-muted inline-flex h-8 items-center rounded-md border px-3 text-sm"
                  >
                    Explorer
                  </Link>
                  <Link
                    to={`/logs?group=${encodeURIComponent(selectedGroup.id)}`}
                    className="border-input bg-background hover:bg-muted inline-flex h-8 items-center rounded-md border px-3 text-sm"
                  >
                    Logs
                  </Link>
                  <Button size="sm" variant="destructive" onClick={() => removeGroup(selectedGroup.id)}>
                    Remover
                  </Button>
                </div>
              </div>

              {summary ? (
                <div className="grid gap-4 sm:grid-cols-3">
                  <Card>
                    <CardContent className="pt-6">
                      <p className="text-muted-foreground text-xs">Membros</p>
                      <p className="text-2xl font-semibold tabular-nums">{summary.member_count}</p>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardContent className="pt-6">
                      <p className="text-muted-foreground text-xs">CPU média</p>
                      <p className="text-2xl font-semibold tabular-nums">
                        {formatPercent(summary.avg_cpu)}
                      </p>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardContent className="pt-6">
                      <p className="text-muted-foreground text-xs">Memória média</p>
                      <p className="text-2xl font-semibold tabular-nums">
                        {formatPercent(summary.avg_memory_pct)}
                      </p>
                      <p className="text-muted-foreground text-xs">
                        {formatBytes(summary.total_memory)} total
                      </p>
                    </CardContent>
                  </Card>
                </div>
              ) : (
                <Skeleton className="h-24 w-full" />
              )}

              <ScrollArea className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Container</TableHead>
                      <TableHead>Stack</TableHead>
                      <TableHead>Serviço</TableHead>
                      <TableHead>CPU</TableHead>
                      <TableHead>Memória</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {(summary?.members ?? []).map((m) => (
                      <TableRow key={m.entity_uid}>
                        <TableCell className="font-medium">{m.container}</TableCell>
                        <TableCell className="text-muted-foreground text-xs">{m.stack || "—"}</TableCell>
                        <TableCell className="text-muted-foreground text-xs">{m.service || "—"}</TableCell>
                        <TableCell>{formatPercent(m.cpu_usage)}</TableCell>
                        <TableCell className="text-sm">{formatBytes(m.memory_usage)}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </ScrollArea>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
