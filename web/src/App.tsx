import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom"

import { AppShell } from "@/components/layout/app-shell"
import { RedirectWithParams } from "@/components/layout/redirect"
import { TooltipProvider } from "@/components/ui/tooltip"
import { ContainersPage } from "@/pages/containers"
import { LogsPage } from "@/pages/logs"
import { MetricsPage } from "@/pages/metrics"
import { OverviewPage } from "@/pages/overview"
import { ProblemsPage } from "@/pages/problems"
import { TopologyPage } from "@/pages/topology"
import { TracesPage } from "@/pages/traces"

export default function App() {
  return (
    <TooltipProvider>
      <BrowserRouter>
        <Routes>
          <Route element={<AppShell />}>
            <Route index element={<OverviewPage />} />
            <Route path="problems" element={<ProblemsPage />} />
            <Route path="containers" element={<ContainersPage />} />
            <Route path="metrics" element={<MetricsPage />} />
            <Route path="logs" element={<LogsPage />} />
            <Route path="traces" element={<TracesPage />} />
            <Route path="topology" element={<TopologyPage />} />

            <Route path="workloads" element={<RedirectWithParams to="/containers" />} />
            <Route path="fleet" element={<RedirectWithParams to="/containers" params={{ view: "table" }} />} />
            <Route path="groups" element={<RedirectWithParams to="/containers" params={{ panel: "groups" }} />} />
            <Route path="explorer" element={<RedirectWithParams to="/metrics" params={{ mode: "compare" }} />} />
            <Route path="patterns" element={<RedirectWithParams to="/logs" params={{ mode: "patterns" }} />} />
            <Route path="insights" element={<RedirectWithParams to="/problems" />} />
            <Route path="slos" element={<RedirectWithParams to="/problems" params={{ tab: "slos" }} />} />
            <Route path="events" element={<RedirectWithParams to="/problems" params={{ tab: "history" }} />} />

            <Route path="*" element={<Navigate to="/" replace />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </TooltipProvider>
  )
}
