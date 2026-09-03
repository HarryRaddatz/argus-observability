import { Activity, Compass, Layers, LayoutDashboard, Lightbulb, ScrollText, Server, Ship, Zap } from "lucide-react"
import { NavLink, useLocation } from "react-router-dom"

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar"

const nav = [
  { to: "/", label: "Visão geral", icon: LayoutDashboard, end: true },
  { to: "/workloads", label: "Workloads", icon: Server, end: false },
  { to: "/fleet", label: "Fleet", icon: Ship, end: false },
  { to: "/groups", label: "Grupos", icon: Layers, end: false },
  { to: "/metrics", label: "Métricas", icon: Activity, end: false },
  { to: "/explorer", label: "Explorer", icon: Compass, end: false },
  { to: "/insights", label: "Insights", icon: Lightbulb, end: false },
  { to: "/logs", label: "Logs", icon: ScrollText, end: false },
  { to: "/events", label: "Eventos", icon: Zap, end: false },
]

export function AppSidebar() {
  const location = useLocation()

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" render={<NavLink to="/" />}>
              <div className="bg-primary text-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg font-semibold">
                A
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-semibold">Argus</span>
                <span className="truncate text-xs text-muted-foreground">Gestão</span>
              </div>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Navegação</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {nav.map((item) => (
                <SidebarMenuItem key={item.to}>
                  <SidebarMenuButton
                    tooltip={item.label}
                    isActive={
                      item.end
                        ? location.pathname === item.to
                        : location.pathname.startsWith(item.to)
                    }
                    render={<NavLink to={item.to} end={item.end} />}
                  >
                    <item.icon />
                    <span>{item.label}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        <p className="text-muted-foreground px-2 text-xs">shadcn/ui</p>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
