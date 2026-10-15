import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { STAT_MODES, statLabel, type StatMode } from "@/lib/stats"

export function StatToggle({ value, onChange }: { value: StatMode; onChange: (mode: StatMode) => void }) {
  const { t } = useTranslation()
  return (
    <div role="group" aria-label={t("chart.summary")} className="flex flex-wrap gap-1">
      {STAT_MODES.map((mode) => (
        <Button
          key={mode}
          size="sm"
          variant={value === mode ? "default" : "outline"}
          aria-pressed={value === mode}
          onClick={() => onChange(mode)}
        >
          {statLabel(t, mode)}
        </Button>
      ))}
    </div>
  )
}
