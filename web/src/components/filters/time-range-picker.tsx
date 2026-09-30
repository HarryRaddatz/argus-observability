import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { TIME_RANGES } from "@/lib/observability"

type Props = {
  value: string
  onChange: (value: string) => void
}

export function TimeRangePicker({ value, onChange }: Props) {
  const { t } = useTranslation()

  return (
    <div role="group" aria-label={t("time.period")} className="flex gap-1">
      {TIME_RANGES.map((r) => (
        <Button
          key={r.id}
          size="sm"
          variant={value === r.id ? "default" : "outline"}
          aria-pressed={value === r.id}
          onClick={() => onChange(r.id)}
        >
          {t(`time.${r.id}`)}
        </Button>
      ))}
    </div>
  )
}
