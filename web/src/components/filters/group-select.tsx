import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { listWorkloadGroups, type WorkloadGroup } from "@/lib/api"

type Props = {
  value: string
  onChange: (value: string) => void
  disabled?: boolean
  groups?: WorkloadGroup[]
}

export function GroupSelect({ value, onChange, disabled, groups: provided }: Props) {
  const { t } = useTranslation()
  const [loaded, setLoaded] = useState<WorkloadGroup[]>([])

  useEffect(() => {
    if (provided) return
    listWorkloadGroups().then(setLoaded).catch(() => setLoaded([]))
  }, [provided])

  const groups = provided ?? loaded

  return (
    <select
      aria-label={t("common.group")}
      className="border-input bg-background h-8 rounded-md border px-2 text-sm disabled:opacity-50"
      value={value}
      disabled={disabled}
      onChange={(e) => onChange(e.target.value)}
    >
      <option value="">{t("common.allGroups")}</option>
      {groups.map((g) => (
        <option key={g.id} value={g.id}>
          {g.name}
        </option>
      ))}
    </select>
  )
}
