import {
  Activity,
  Compass,
  GitBranch,
  GitCommitHorizontal,
  LayoutDashboard,
  Layers,
  Lightbulb,
  Repeat,
  ScrollText,
  Server,
  Ship,
  Target,
  Zap,
  type LucideIcon,
} from "lucide-react"

export type NavItem = {
  to: string
  label: string
  icon: LucideIcon
  end?: boolean
  description?: string
}

export type NavGroup = {
  label: string
  items: NavItem[]
}

export const navGroups: NavGroup[] = [
  {
    label: "Visão",
    items: [
      {
        to: "/",
        label: "Dashboard",
        icon: LayoutDashboard,
        end: true,
        description: "Resumo de infraestrutura, HTTP e alertas",
      },
    ],
  },
  {
    label: "Infraestrutura",
    items: [
      { to: "/workloads", label: "Workloads", icon: Server, description: "Containers monitorados pelo agent" },
      { to: "/fleet", label: "Fleet", icon: Ship, description: "Estado operacional e restarts" },
      { to: "/groups", label: "Grupos", icon: Layers, description: "Agrupamentos por stack ou serviço" },
    ],
  },
  {
    label: "Telemetria",
    items: [
      { to: "/metrics", label: "Métricas", icon: Activity, description: "Detalhe por container" },
      { to: "/explorer", label: "Explorer", icon: Compass, description: "Comparar séries entre containers" },
      { to: "/logs", label: "Logs", icon: ScrollText, description: "Busca e stream de logs" },
    ],
  },
  {
    label: "Análise",
    items: [
      { to: "/insights", label: "Insights", icon: Lightbulb, description: "Achados automáticos" },
      { to: "/patterns", label: "Patterns", icon: Repeat, description: "Padrões de log" },
      { to: "/topology", label: "Topologia", icon: GitBranch, description: "Grafo de dependências" },
      { to: "/traces", label: "Traces", icon: GitCommitHorizontal, description: "Rastreamento distribuído" },
      { to: "/slos", label: "SLOs", icon: Target, description: "Objetivos de nível de serviço" },
    ],
  },
  {
    label: "Alertas",
    items: [
      { to: "/events", label: "Eventos", icon: Zap, description: "Timeline de eventos e regras" },
    ],
  },
]

export function findRouteMeta(pathname: string): { title: string; group?: string; description?: string } {
  for (const group of navGroups) {
    for (const item of group.items) {
      const active = item.end ? pathname === item.to : pathname.startsWith(item.to)
      if (active) {
        return { title: item.label, group: group.label, description: item.description }
      }
    }
  }
  return { title: "Argus" }
}
