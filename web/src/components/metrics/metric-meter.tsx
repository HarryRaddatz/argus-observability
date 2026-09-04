import { cn } from "@/lib/utils"
import { formatPercent } from "@/lib/format"

type Props = {
  label: string
  value: number
  warnAt?: number
  criticalAt?: number
  className?: string
}

export function MetricMeter({ label, value, warnAt = 75, criticalAt = 90, className }: Props) {
  const pct = Math.min(100, Math.max(0, value))
  return (
    <div className={cn("space-y-1", className)}>
      <div className="flex items-center justify-between gap-2 text-xs">
        <span className="text-muted-foreground">{label}</span>
        <span
          className={cn(
            "font-mono tabular-nums",
            pct >= criticalAt && "text-destructive",
            pct >= warnAt && pct < criticalAt && "text-amber-600",
          )}
        >
          {formatPercent(pct)}
        </span>
      </div>
      <div className="bg-muted h-2 overflow-hidden rounded-full">
        <div
          className={cn(
            "h-full transition-all",
            pct >= criticalAt ? "bg-destructive" : pct >= warnAt ? "bg-amber-500" : "bg-primary",
          )}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  )
}
