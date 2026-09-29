import { useCallback, useEffect, useState } from "react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet"
import {
  createWorkloadGroup,
  deleteWorkloadGroup,
  discoverWorkloadGroups,
  type WorkloadGroup,
  type WorkloadGroupInput,
} from "@/lib/api"
import { cn } from "@/lib/utils"

const kindLabel: Record<string, string> = {
  stack: "Stack",
  service: "Serviço",
  custom: "Personalizado",
}

function membersLabel(n: number) {
  return n === 1 ? "1 container" : `${n} containers`
}

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  groups: WorkloadGroup[]
  containers: string[]
  activeGroup: string
  onSelectGroup: (id: string) => void
  onChanged: () => void
}

export function GroupsPanel({ open, onOpenChange, groups, containers, activeGroup, onSelectGroup, onChanged }: Props) {
  const [discovered, setDiscovered] = useState<WorkloadGroup[]>([])
  const [customName, setCustomName] = useState("")
  const [customSelected, setCustomSelected] = useState<string[]>([])
  const [error, setError] = useState<string | null>(null)

  const loadDiscovered = useCallback(() => {
    discoverWorkloadGroups().then(setDiscovered).catch(() => setDiscovered([]))
  }, [])

  useEffect(() => {
    if (open) loadDiscovered()
  }, [open, loadDiscovered])

  const savedKeys = new Set(groups.map((g) => `${g.kind}:${g.label_value ?? g.name}`))
  const suggestions = discovered.filter((g) => !savedKeys.has(`${g.kind}:${g.label_value ?? g.name}`)).slice(0, 8)

  async function run(action: () => Promise<unknown>) {
    try {
      await action()
      setError(null)
      onChanged()
      loadDiscovered()
    } catch (e) {
      setError(e instanceof Error ? e.message : "Não foi possível salvar")
    }
  }

  function importDiscovered(g: WorkloadGroup) {
    const input: WorkloadGroupInput = {
      name: g.name,
      kind: g.kind,
      label_key: g.label_key,
      label_value: g.label_value,
      description: `Importado de ${kindLabel[g.kind] ?? g.kind} ${g.label_value ?? ""}`.trim(),
    }
    return run(() => createWorkloadGroup(input))
  }

  function saveCustomGroup() {
    if (!customName.trim() || customSelected.length === 0) return
    return run(async () => {
      await createWorkloadGroup({
        name: customName.trim(),
        kind: "custom",
        containers: customSelected,
        description: "Grupo personalizado",
      })
      setCustomName("")
      setCustomSelected([])
    })
  }

  function removeGroup(id: string) {
    if (activeGroup === id) onSelectGroup("")
    return run(() => deleteWorkloadGroup(id))
  }

  function toggleCustomContainer(name: string) {
    setCustomSelected((prev) => (prev.includes(name) ? prev.filter((c) => c !== name) : [...prev, name]))
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-full overflow-y-auto sm:max-w-md">
        <SheetHeader>
          <SheetTitle>Grupos</SheetTitle>
          <SheetDescription>Use um grupo para filtrar Containers, Métricas, Logs e Problemas.</SheetDescription>
        </SheetHeader>

        <div className="space-y-6 px-4 pb-6">
          {error ? <p className="text-destructive text-sm">{error}</p> : null}

          <section className="space-y-2">
            <h2 className="text-sm font-medium">Seus grupos</h2>
            {groups.length === 0 ? (
              <p className="text-muted-foreground text-sm">Nenhum grupo salvo. Adicione um sugerido ou crie um abaixo.</p>
            ) : (
              <ul className="space-y-1">
                {groups.map((g) => (
                  <li
                    key={g.id}
                    className={cn(
                      "flex items-center justify-between gap-2 rounded-md border px-3 py-2",
                      activeGroup === g.id && "border-primary/40 bg-muted",
                    )}
                  >
                    <button type="button" className="min-w-0 flex-1 text-left" onClick={() => onSelectGroup(g.id)}>
                      <span className="block truncate text-sm font-medium">{g.name}</span>
                      <span className="text-muted-foreground block text-xs">{membersLabel(g.member_count ?? 0)}</span>
                    </button>
                    <Badge variant="outline">{kindLabel[g.kind] ?? g.kind}</Badge>
                    <Button size="sm" variant="ghost" onClick={() => removeGroup(g.id)} aria-label={`Remover o grupo ${g.name}`}>
                      Remover
                    </Button>
                  </li>
                ))}
              </ul>
            )}
          </section>

          {suggestions.length > 0 ? (
            <section className="space-y-2">
              <h2 className="text-sm font-medium">Sugeridos pelos labels dos containers</h2>
              <ul className="space-y-2">
                {suggestions.map((g) => (
                  <li key={g.id} className="flex items-center justify-between gap-2 text-sm">
                    <div className="min-w-0">
                      <p className="truncate font-medium">{g.name}</p>
                      <p className="text-muted-foreground text-xs">{membersLabel(g.member_count ?? 0)}</p>
                    </div>
                    <Button size="sm" variant="outline" onClick={() => importDiscovered(g)}>
                      Adicionar
                    </Button>
                  </li>
                ))}
              </ul>
            </section>
          ) : null}

          <section className="space-y-3">
            <h2 className="text-sm font-medium">Novo grupo</h2>
            <Input
              aria-label="Nome do grupo"
              placeholder="Nome do grupo"
              value={customName}
              onChange={(e) => setCustomName(e.target.value)}
            />
            <ScrollArea className="h-40 rounded-md border">
              <ul className="p-1">
                {containers.map((c) => (
                  <li key={c}>
                    <button
                      type="button"
                      aria-pressed={customSelected.includes(c)}
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
            <Button
              size="sm"
              className="w-full"
              disabled={!customName.trim() || customSelected.length === 0}
              onClick={saveCustomGroup}
            >
              {customSelected.length === 0
                ? "Selecione containers para criar"
                : `Criar grupo com ${membersLabel(customSelected.length)}`}
            </Button>
          </section>
        </div>
      </SheetContent>
    </Sheet>
  )
}
