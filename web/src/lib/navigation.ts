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
  labelKey: string
  icon: LucideIcon
  end?: boolean
}

export const navItems: NavItem[] = [
  { to: "/", labelKey: "nav.overview", icon: LayoutDashboard, end: true },
  { to: "/problems", labelKey: "nav.problems", icon: TriangleAlert },
  { to: "/containers", labelKey: "nav.containers", icon: Boxes },
  { to: "/metrics", labelKey: "nav.metrics", icon: Activity },
  { to: "/logs", labelKey: "nav.logs", icon: ScrollText },
  { to: "/traces", labelKey: "nav.traces", icon: GitCommitHorizontal },
  { to: "/topology", labelKey: "nav.topology", icon: GitBranch },
]

export function isNavItemActive(item: NavItem, pathname: string) {
  return item.end ? pathname === item.to : pathname.startsWith(item.to)
}
