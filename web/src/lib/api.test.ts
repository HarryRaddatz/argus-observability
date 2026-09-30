import { afterEach, describe, expect, it, vi } from "vitest"

import { fetchMetricSeries, getHealth, searchLogs } from "@/lib/api"

function jsonResponse(data: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: status === 200 ? "OK" : "Error",
    json: async () => data,
  } as Response
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe("api client", () => {
  it("builds metric series query params", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ metric_name: "cpu.usage", series: [] }))
    vi.stubGlobal("fetch", fetchMock)
    await fetchMetricSeries("cpu.usage", "6h", "api", "g1")
    expect(fetchMock).toHaveBeenCalledTimes(1)
    const url = String(fetchMock.mock.calls[0][0])
    expect(url).toContain("/api/v1/metrics/series?")
    expect(url).toContain("metric=cpu.usage")
    expect(url).toContain("since=6h")
    expect(url).toContain("container=api")
    expect(url).toContain("group=g1")
  })

  it("omits all-sentinels from log search", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse([]))
    vi.stubGlobal("fetch", fetchMock)
    await searchLogs({ since: "1h", level: "all", container: "all", topic: "all", q: "oom" })
    const url = String(fetchMock.mock.calls[0][0])
    expect(url).toContain("q=oom")
    expect(url).not.toContain("level=")
    expect(url).not.toContain("container=")
    expect(url).not.toContain("topic=")
  })

  it("throws on non-OK health", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ status: "down" }, 503)))
    await expect(getHealth()).rejects.toThrow(/503/)
  })
})
