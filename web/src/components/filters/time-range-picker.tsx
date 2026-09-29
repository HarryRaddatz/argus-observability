import { Button } from "@/components/ui/button"
import { TIME_RANGES } from "@/lib/observability"

type Props = {
  value: string
  onChange: (value: string) => void
}

export function TimeRangePicker({ value, onChange }: Props) {
  return (
    <div role="group" aria-label="Período" className="flex gap-1">
      {TIME_RANGES.map((r) => (
        <Button
          key={r.id}
          size="sm"
          variant={value === r.id ? "default" : "outline"}
          aria-pressed={value === r.id}
          onClick={() => onChange(r.id)}
        >
          {r.label}
        </Button>
      ))}
    </div>
  )
}
