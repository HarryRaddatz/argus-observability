import type { TFunction } from "i18next"

export const severityVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  critical: "destructive",
  warning: "secondary",
  info: "outline",
  debug: "outline",
}

export function severityLabel(t: TFunction, severity: string): string {
  switch (severity) {
    case "critical":
      return t("severity.critical")
    case "warning":
      return t("severity.warning")
    case "info":
      return t("severity.info")
    case "debug":
      return t("severity.debug")
    default:
      return severity
  }
}
