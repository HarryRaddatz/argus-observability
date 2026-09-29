import {
  Activity,
  Boxes,
  GitBranch,
  GitCommitHorizontal,
  LayoutDashboard,
  ScrollText,
  TriangleAlert,
  type LucideIcon,
} from "lucide-react"

export type NavItem = {
  to: string
  label: string
  icon: LucideIcon
  end?: boolean
}

export const navItems: NavItem[] = [
  { to: "/", label: "Visão geral", icon: LayoutDashboard, end: true },
  { to: "/problems", label: "Problemas", icon: TriangleAlert },
  { to: "/containers", label: "Containers", icon: Boxes },
  { to: "/metrics", label: "Métricas", icon: Activity },
  { to: "/logs", label: "Logs", icon: ScrollText },
  { to: "/traces", label: "Traces", icon: GitCommitHorizontal },
  { to: "/topology", label: "Topologia", icon: GitBranch },
]

export function isNavItemActive(item: NavItem, pathname: string) {
  return item.end ? pathname === item.to : pathname.startsWith(item.to)
}
