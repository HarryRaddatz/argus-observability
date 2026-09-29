export type MeterTone = "default" | "warning" | "critical"

export function meterTone(value: number, warnAt = 75, criticalAt = 90): MeterTone {
  if (value >= criticalAt) return "critical"
  if (value >= warnAt) return "warning"
  return "default"
}

export const barTone: Record<MeterTone, string> = {
  default: "bg-primary",
  warning: "bg-amber-500",
  critical: "bg-destructive",
}

export const textTone: Record<MeterTone, string> = {
  default: "",
  warning: "text-amber-600",
  critical: "text-destructive",
}
