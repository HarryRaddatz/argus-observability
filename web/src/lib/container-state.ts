import type { ContainerFleetStatus } from "@/lib/api"

type BadgeVariant = "default" | "secondary" | "destructive" | "outline"

const variants: Record<string, BadgeVariant> = {
  running: "default",
  restarting: "secondary",
  exited: "outline",
  dead: "destructive",
}

const labels: Record<string, string> = {
  running: "Rodando",
  restarting: "Reiniciando",
  exited: "Parado",
  dead: "Morto",
  paused: "Pausado",
  created: "Criado",
}

export function stateVariant(state: string): BadgeVariant {
  return variants[state] ?? "outline"
}

export function stateLabel(state: string): string {
  return labels[state] ?? state
}

export function isUnstable(status: ContainerFleetStatus): boolean {
  const state = status.state?.toLowerCase() ?? ""
  return (
    state === "restarting" ||
    state === "dead" ||
    status.health === "unhealthy" ||
    Boolean(status.oom_killed) ||
    status.restart_count > 3
  )
}

export function instabilityReason(status: ContainerFleetStatus): string {
  const state = status.state?.toLowerCase() ?? ""
  if (status.oom_killed) return "Encerrado por falta de memória"
  if (state === "dead") return "Container morto"
  if (state === "restarting") return "Reiniciando agora"
  if (status.health === "unhealthy") return "Healthcheck falhando"
  return `${status.restart_count} reinícios`
}
