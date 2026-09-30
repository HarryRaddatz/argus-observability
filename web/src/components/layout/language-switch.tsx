import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import type { AppLocale } from "@/i18n"

export function LanguageSwitch() {
  const { t, i18n } = useTranslation()
  const current = i18n.language.startsWith("pt") ? "pt-BR" : "en"

  const setLocale = (lng: AppLocale) => {
    void i18n.changeLanguage(lng)
  }

  return (
    <div role="group" aria-label={t("shell.language")} className="flex gap-1">
      <Button size="sm" variant={current === "en" ? "default" : "outline"} aria-pressed={current === "en"} onClick={() => setLocale("en")}>
        {t("shell.en")}
      </Button>
      <Button size="sm" variant={current === "pt-BR" ? "default" : "outline"} aria-pressed={current === "pt-BR"} onClick={() => setLocale("pt-BR")}>
        {t("shell.pt")}
      </Button>
    </div>
  )
}
