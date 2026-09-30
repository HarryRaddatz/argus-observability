import { Outlet } from "react-router-dom"
import { useTranslation } from "react-i18next"

import { AppSidebar } from "@/components/layout/app-sidebar"
import { LanguageSwitch } from "@/components/layout/language-switch"
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar"

export function AppShell() {
  const { t } = useTranslation()

  return (
    <SidebarProvider>
      <a
        href="#main"
        className="bg-background focus-visible:ring-ring sr-only z-50 rounded-md px-3 py-2 text-sm font-medium focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus-visible:ring-2"
      >
        {t("shell.skip")}
      </a>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-12 shrink-0 items-center justify-between gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" aria-label={t("shell.menu")} />
          <LanguageSwitch />
        </header>
        <main id="main" tabIndex={-1} className="flex flex-1 flex-col gap-4 p-4 outline-none md:p-6">
          <Outlet />
        </main>
      </SidebarInset>
    </SidebarProvider>
  )
}
