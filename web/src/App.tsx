import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom"

import { AppShell } from "@/components/layout/app-shell"
import { TooltipProvider } from "@/components/ui/tooltip"
import { DashboardPage } from "@/pages/dashboard"
import { EventsPage } from "@/pages/events"
import { LogsPage } from "@/pages/logs"
import { MetricsPage } from "@/pages/metrics"
import { WorkloadsPage } from "@/pages/workloads"

export default function App() {
  return (
    <TooltipProvider>
      <BrowserRouter>
        <Routes>
          <Route element={<AppShell />}>
            <Route index element={<DashboardPage />} />
            <Route path="workloads" element={<WorkloadsPage />} />
            <Route path="metrics" element={<MetricsPage />} />
            <Route path="logs" element={<LogsPage />} />
            <Route path="events" element={<EventsPage />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </TooltipProvider>
  )
}
