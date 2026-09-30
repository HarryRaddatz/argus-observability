import type { TFunction } from "i18next"

import type { ContainerFleetStatus } from "@/lib/api"

type BadgeVariant = "default" | "secondary" | "destructive" | "outline"

const variants: Record<string, BadgeVariant> = {
  running: "default",
  restarting: "secondary",
  exited: "outline",
  dead: "destructive",
}

export function stateVariant(state: string): BadgeVariant {
  return variants[state] ?? "outline"
}

export function stateLabel(t: TFunction, state: string): string {
  const key = `state.${state}`
  const translated = t(key)
  return translated === key ? state : translated
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

export function instabilityReason(t: TFunction, status: ContainerFleetStatus): string {
  const state = status.state?.toLowerCase() ?? ""
  if (status.oom_killed) return t("unstable.oom")
  if (state === "dead") return t("unstable.dead")
  if (state === "restarting") return t("unstable.restarting")
  if (status.health === "unhealthy") return t("unstable.unhealthy")
  return t("unstable.restarts", { count: status.restart_count })
}
