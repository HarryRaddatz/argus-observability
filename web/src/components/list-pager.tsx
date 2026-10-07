import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"

type Props = {
  start: number
  end: number
  total: number
  page: number
  pageSize: number
  truncated?: boolean
  onPage: (page: number) => void
}

export function ListPager({ start, end, total, page, pageSize, truncated, onPage }: Props) {
  const { t } = useTranslation()
  const pages = Math.max(1, Math.ceil(total / pageSize))
  const safe = Math.min(page, pages)
  return (
    <div className="flex flex-wrap items-center justify-between gap-3">
      <p className="text-muted-foreground text-sm tabular-nums">
        {t("common.pagerRange", { start, end, total })}
        {truncated ? ` ${t("common.pagerTruncated")}` : ""}
      </p>
      <div className="flex gap-2">
        <Button type="button" size="sm" variant="outline" disabled={safe <= 1} onClick={() => onPage(safe - 1)}>
          {t("common.pagerPrev")}
        </Button>
        <Button type="button" size="sm" variant="outline" disabled={safe >= pages || total === 0} onClick={() => onPage(safe + 1)}>
          {t("common.pagerNext")}
        </Button>
      </div>
    </div>
  )
}
