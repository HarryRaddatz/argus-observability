import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom"

import { AppShell } from "@/components/layout/app-shell"
import { TooltipProvider } from "@/components/ui/tooltip"
import { DashboardPage } from "@/pages/dashboard"
import { EventsPage } from "@/pages/events"
import { ExplorerPage } from "@/pages/explorer"
import { InsightsPage } from "@/pages/insights"
import { LogsPage } from "@/pages/logs"
import { MetricsPage } from "@/pages/metrics"
import { FleetPage } from "@/pages/fleet"
import { GroupsPage } from "@/pages/groups"
import { PatternsPage } from "@/pages/patterns"
import { TopologyPage } from "@/pages/topology"
import { TracesPage } from "@/pages/traces"
import { SLOsPage } from "@/pages/slos"
import { WorkloadsPage } from "@/pages/workloads"

export default function App() {
  return (
    <TooltipProvider>
      <BrowserRouter>
        <Routes>
          <Route element={<AppShell />}>
            <Route index element={<DashboardPage />} />
            <Route path="workloads" element={<WorkloadsPage />} />
            <Route path="fleet" element={<FleetPage />} />
            <Route path="groups" element={<GroupsPage />} />
            <Route path="metrics" element={<MetricsPage />} />
            <Route path="explorer" element={<ExplorerPage />} />
            <Route path="insights" element={<InsightsPage />} />
            <Route path="patterns" element={<PatternsPage />} />
            <Route path="topology" element={<TopologyPage />} />
            <Route path="traces" element={<TracesPage />} />
            <Route path="slos" element={<SLOsPage />} />
            <Route path="logs" element={<LogsPage />} />
            <Route path="events" element={<EventsPage />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </TooltipProvider>
  )
}
