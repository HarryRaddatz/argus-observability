import { Outlet, useLocation } from "react-router-dom"

import { AppSidebar } from "@/components/layout/app-sidebar"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { findRouteMeta } from "@/lib/navigation"

export function AppShell() {
  const location = useLocation()
  const meta = findRouteMeta(location.pathname)

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-14 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <div className="min-w-0">
            {meta.group ? (
              <p className="text-muted-foreground truncate text-xs">{meta.group}</p>
            ) : null}
            <p className="truncate text-sm font-medium">{meta.title}</p>
          </div>
        </header>
        <main className="flex flex-1 flex-col gap-4 p-4 md:p-6">
          <Outlet />
        </main>
      </SidebarInset>
    </SidebarProvider>
  )
}
