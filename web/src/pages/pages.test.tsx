import { render, screen, waitFor } from "@testing-library/react"
import { MemoryRouter } from "react-router-dom"
import { afterEach, describe, expect, it, vi } from "vitest"

import { ContainersPage } from "@/pages/containers"
import { LogsPage } from "@/pages/logs"
import { OverviewPage } from "@/pages/overview"
import { TracesPage } from "@/pages/traces"

const emptyFleet = {
  updated_at: "2026-01-01T00:00:00Z",
  summary: {
    running: 0,
    exited: 0,
    restarting: 0,
    unhealthy: 0,
    dead: 0,
    total_restart_count: 0,
    replicas_up: 0,
    replicas_total: 0,
  },
  services: [],
  containers: [] as { container: string; entity_uid: string; state: string; restart_count: number }[],
  events_24h: { restarts_24h: 0, failures_24h: 0, oom_24h: 0, disconnect_24h: 0 },
}

function jsonResponse(data: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: status === 200 ? "OK" : "Error",
    json: async () => data,
  } as Response
}

function mockHub(routes: Record<string, unknown>) {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: RequestInfo | URL) => {
      const raw = String(input)
      const path = raw.split("?")[0]
      if (path in routes) return jsonResponse(routes[path])
      const key = Object.keys(routes).find((k) => path.endsWith(k) || raw.includes(k))
      if (key) return jsonResponse(routes[key])
      return jsonResponse({}, 404)
    }),
  )
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe("panel pages", () => {
  it("renders overview empty state after health succeeds", async () => {
    mockHub({
      "/health": { status: "ok" },
      "/api/v1/alerts/active": [],
      "/api/v1/slos/status": [],
      "/api/v1/fleet/status": emptyFleet,
      "/api/v1/metrics/http/summary": [],
    })
    render(
      <MemoryRouter>
        <OverviewPage />
      </MemoryRouter>,
    )
    expect(screen.getByRole("heading", { name: "Overview" })).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.getByText("Nothing needs action right now.")).toBeInTheDocument()
    })
  })

  it("renders containers heading from fleet and workloads", async () => {
    mockHub({
      "/api/v1/workloads": [
        {
          container: "api",
          entity_uid: "docker:host:api",
          cpu_usage: 1,
          memory_usage: 1,
          memory_limit: 10,
          updated_at: "2026-01-01T00:00:00Z",
        },
      ],
      "/api/v1/fleet/status": {
        ...emptyFleet,
        containers: [{ container: "api", entity_uid: "docker:host:api", state: "running", restart_count: 0 }],
      },
      "/api/v1/metrics/series": { metric_name: "cpu.usage", series: [] },
      "/api/v1/workload-groups": [],
    })
    render(
      <MemoryRouter>
        <ContainersPage />
      </MemoryRouter>,
    )
    expect(screen.getByRole("heading", { name: "Containers" })).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.getAllByText("api").length).toBeGreaterThan(0)
    })
  })

  it("shows log filters and a clickable trace id", async () => {
    mockHub({
      "/api/v1/workloads": [],
      "/api/v1/workload-groups": [],
      "/api/v1/logs/search": [
        {
          ts: "2026-01-01T00:00:00Z",
          message: "request failed",
          level: "error",
          entity_uid: "docker:host:api",
          labels: {},
          fields: { trace_id: "abc123def4567890", topics: ["errors"] },
        },
      ],
    })
    render(
      <MemoryRouter>
        <LogsPage />
      </MemoryRouter>,
    )
    expect(screen.getByRole("heading", { name: "Logs" })).toBeInTheDocument()
    expect(screen.getByLabelText("Container")).toBeInTheDocument()
    expect(screen.getByLabelText("Trace ID")).toBeInTheDocument()
    await waitFor(() => {
      const link = screen.getByRole("link", { name: "abc123de…" })
      expect(link).toHaveAttribute("href", "/traces?trace_id=abc123def4567890")
    })
  })

  it("renders traces list from the hub", async () => {
    mockHub({
      "/api/v1/traces": [
        {
          trace_id: "trace-1",
          source: "otlp",
          service: "checkout",
          container: "api",
          name: "POST /pay",
          start_ts: "2026-01-01T00:00:00Z",
          end_ts: "2026-01-01T00:00:01Z",
          duration_ms: 12,
          span_count: 2,
          error: false,
        },
      ],
    })
    render(
      <MemoryRouter>
        <TracesPage />
      </MemoryRouter>,
    )
    expect(screen.getByRole("heading", { name: "Traces" })).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.getByText("POST /pay")).toBeInTheDocument()
    })
  })
})
