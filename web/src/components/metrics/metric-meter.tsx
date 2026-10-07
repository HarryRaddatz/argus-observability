import { formatPercent } from "@/lib/format"
import { barTone, meterTone, textTone, type MeterTone } from "@/lib/meter"
import { cn } from "@/lib/utils"

export function UsageBar({ value, tone = "default", label }: { value: number; tone?: MeterTone; label: string }) {
  const pct = Math.min(100, Math.max(0, value))
  return (
    <div
      role="meter"
      aria-label={label}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-valuenow={Math.round(pct)}
      className="bg-muted h-2 overflow-hidden rounded-full"
    >
      <div className={cn("h-full transition-all", barTone[tone])} style={{ width: `${pct}%` }} />
    </div>
  )
}

type Props = {
  label: string
  value: number
  warnAt?: number
  criticalAt?: number
  className?: string
}

export function MetricMeter({ label, value, warnAt = 75, criticalAt = 90, className }: Props) {
  const pct = Math.max(0, value)
  const tone = meterTone(Math.min(100, pct), warnAt, criticalAt)
  return (
    <div className={cn("space-y-1", className)}>
      <div className="flex items-center justify-between gap-2 text-xs">
        <span className="text-muted-foreground">{label}</span>
        <span className={cn("font-mono tabular-nums", textTone[tone])}>{formatPercent(pct)}</span>
      </div>
      <UsageBar value={pct} tone={tone} label={label} />
    </div>
  )
}
