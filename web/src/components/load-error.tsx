import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"

export function LoadError({ message, onRetry }: { message: string; onRetry?: () => void }) {
  const { t } = useTranslation()
  return (
    <div role="alert" className="border-destructive/40 bg-destructive/5 flex flex-wrap items-center gap-3 rounded-md border px-3 py-2 text-sm">
      <span className="text-destructive">{message}</span>
      {onRetry ? (
        <Button size="sm" variant="outline" onClick={onRetry}>
          {t("common.retry")}
        </Button>
      ) : null}
    </div>
  )
}
