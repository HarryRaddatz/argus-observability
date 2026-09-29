export const severityVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  critical: "destructive",
  warning: "secondary",
  info: "outline",
  debug: "outline",
}

export const severityLabel: Record<string, string> = {
  critical: "Crítico",
  warning: "Atenção",
  info: "Informativo",
  debug: "Depuração",
}
