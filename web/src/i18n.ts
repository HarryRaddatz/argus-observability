import i18n from "i18next"
import { initReactI18next } from "react-i18next"

import en from "./locales/en.json"
import ptBR from "./locales/pt-BR.json"

export const LOCALE_KEY = "argus.locale"
export const locales = ["en", "pt-BR"] as const
export type AppLocale = (typeof locales)[number]

export function isAppLocale(value: string): value is AppLocale {
  return value === "en" || value === "pt-BR"
}

export function readStoredLocale(): AppLocale {
  try {
    const raw = localStorage.getItem(LOCALE_KEY)
    if (raw && isAppLocale(raw)) return raw
  } catch {
    /* ignore */
  }
  return "en"
}

export function localeBcp47(lng: string): string {
  return lng === "pt-BR" ? "pt-BR" : "en"
}

function applyHtmlLang(lng: string) {
  document.documentElement.lang = localeBcp47(lng)
}

void i18n.use(initReactI18next).init({
  resources: {
    en: { translation: en },
    "pt-BR": { translation: ptBR },
  },
  lng: readStoredLocale(),
  fallbackLng: "en",
  interpolation: { escapeValue: false },
})

i18n.on("languageChanged", (lng) => {
  try {
    localStorage.setItem(LOCALE_KEY, isAppLocale(lng) ? lng : "en")
  } catch {
    /* ignore */
  }
  applyHtmlLang(lng)
})

applyHtmlLang(i18n.language)

export default i18n
