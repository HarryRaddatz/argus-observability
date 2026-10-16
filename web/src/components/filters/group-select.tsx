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
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    if (provided) return
    listWorkloadGroups()
      .then((g) => {
        setLoaded(g)
        setFailed(false)
      })
      .catch(() => {
        setLoaded([])
        setFailed(true)
      })
  }, [provided])

  const groups = provided ?? loaded
  const unknown = value && !groups.some((g) => g.id === value)

  return (
    <select
      name="group"
      aria-label={t("common.group")}
      aria-invalid={failed || undefined}
      title={failed ? t("common.groupsError") : undefined}
      className="border-input bg-background text-foreground h-8 rounded-md border px-2 text-sm disabled:opacity-50 aria-invalid:border-destructive pointer-coarse:h-11"
      value={value}
      disabled={disabled}
      onChange={(e) => onChange(e.target.value)}
    >
      <option value="">{t("common.allGroups")}</option>
      {failed ? (
        <option value="" disabled>
          {t("common.groupsError")}
        </option>
      ) : null}
      {unknown ? <option value={value}>{value}</option> : null}
      {groups.map((g) => (
        <option key={g.id} value={g.id}>
          {g.name}
        </option>
      ))}
    </select>
  )
}
