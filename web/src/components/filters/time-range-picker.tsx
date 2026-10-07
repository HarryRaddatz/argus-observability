import { useEffect, useState } from "react"
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

  useEffect(() => {
    const parsed = splitRange(value)
    if (!parsed) return
    setStart(toLocalInput(parsed.start))
    setEnd(toLocalInput(parsed.end))
    setDrafting(true)
  }, [value])

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
            variant={!custom && value === r.id ? "default" : "outline"}
            aria-pressed={!custom && value === r.id}
            onClick={() => {
              setDrafting(false)
              setInvalid(false)
              onChange(r.id)
            }}
          >
            {t(`time.${r.id}`)}
          </Button>
        ))}
        <Button size="sm" variant={custom || drafting ? "default" : "outline"} aria-pressed={custom || drafting} onClick={openCustom}>
          {t("time.custom")}
        </Button>
      </div>
      {drafting || custom ? (
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
              aria-label={t("time.from")}
              value={start}
              onChange={(e) => setStart(e.target.value)}
              className="w-48"
            />
          </label>
          <label className="grid gap-1 text-xs">
            <span className="text-muted-foreground">{t("time.to")}</span>
            <Input
              type="datetime-local"
              aria-label={t("time.to")}
              value={end}
              onChange={(e) => setEnd(e.target.value)}
              className="w-48"
            />
          </label>
          <Button type="submit" size="sm" variant="outline">
            {t("time.apply")}
          </Button>
          {invalid ? <p className="text-destructive text-xs">{t("time.rangeInvalid")}</p> : null}
        </form>
      ) : null}
    </div>
  )
}
