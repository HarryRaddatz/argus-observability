import { useEffect, useId, useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { TIME_RANGES } from "@/lib/observability"
import { defaultLocalRange, isCustomRange, rangeFromLocal, splitRange, toLocalInput } from "@/lib/time-range"

type Props = {
  value: string
  onChange: (value: string) => void
}

export function TimeRangePicker({ value, onChange }: Props) {
  const { t } = useTranslation()
  const custom = isCustomRange(value)
  const [drafting, setDrafting] = useState(custom)
  const [start, setStart] = useState("")
  const [end, setEnd] = useState("")
  const [invalid, setInvalid] = useState(false)
  const errorId = useId()

  useEffect(() => {
    setInvalid(false)
    const parsed = splitRange(value)
    if (!parsed) {
      setDrafting(false)
      return
    }
    setStart(toLocalInput(parsed.start))
    setEnd(toLocalInput(parsed.end))
    setDrafting(true)
  }, [value])

  const customActive = custom || drafting

  function openCustom() {
    if (!start || !end) {
      const next = defaultLocalRange()
      setStart(next.start)
      setEnd(next.end)
    }
    setInvalid(false)
    setDrafting(true)
  }

  function applyCustom() {
    const next = rangeFromLocal(start, end)
    if (!next) {
      setInvalid(true)
      return
    }
    setInvalid(false)
    onChange(next)
  }

  return (
    <div className="flex flex-col items-end gap-2">
      <div role="group" aria-label={t("time.period")} className="flex flex-wrap gap-1">
        {TIME_RANGES.map((r) => (
          <Button
            key={r.id}
            size="sm"
            variant={!customActive && value === r.id ? "default" : "outline"}
            aria-pressed={!customActive && value === r.id}
            onClick={() => {
              setDrafting(false)
              setInvalid(false)
              onChange(r.id)
            }}
          >
            {t(`time.${r.id}`)}
          </Button>
        ))}
        <Button size="sm" variant={customActive ? "default" : "outline"} aria-pressed={customActive} onClick={openCustom}>
          {t("time.custom")}
        </Button>
      </div>
      {customActive ? (
        <form
          className="flex flex-wrap items-end gap-2"
          onSubmit={(e) => {
            e.preventDefault()
            applyCustom()
          }}
        >
          <label className="grid gap-1 text-xs">
            <span className="text-muted-foreground">{t("time.from")}</span>
            <Input
              type="datetime-local"
              name="range_start"
              aria-invalid={invalid || undefined}
              aria-describedby={invalid ? errorId : undefined}
              value={start}
              onChange={(e) => setStart(e.target.value)}
              className="w-48"
            />
          </label>
          <label className="grid gap-1 text-xs">
            <span className="text-muted-foreground">{t("time.to")}</span>
            <Input
              type="datetime-local"
              name="range_end"
              aria-invalid={invalid || undefined}
              aria-describedby={invalid ? errorId : undefined}
              value={end}
              onChange={(e) => setEnd(e.target.value)}
              className="w-48"
            />
          </label>
          <Button type="submit" size="sm" variant="outline">
            {t("time.apply")}
          </Button>
          {invalid ? (
            <p id={errorId} role="alert" className="text-destructive text-xs">
              {t("time.rangeInvalid")}
            </p>
          ) : null}
        </form>
      ) : null}
    </div>
  )
}
