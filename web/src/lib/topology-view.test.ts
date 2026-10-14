import { describe, expect, it } from "vitest"

import { buildServices, filterServices, isAddress, pickService, rankServices, splitAddresses } from "@/lib/topology-view"
import type { TopologySignals } from "@/lib/topology-view"

const signals: TopologySignals = {
  fleet: [
    { container: "shop-payments-1", service: "payments", state: "dead", restart_count: 0, entity_uid: "c1", disposition: "unexpected" },
    { container: "shop-gateway-1", service: "gateway", state: "running", restart_count: 0, entity_uid: "c2" },
  ],
  workloads: [
    { container: "shop-payments-1", service: "payments", entity_uid: "c1", cpu_usage: 80, memory_usage: 100, memory_limit: 200, updated_at: "" },
  ],
  http: [{ service: "gateway", requests: 10, errors: 0, error_rate: 0, avg_latency_ms: 12, max_latency_ms: 40 }],
  alerts: [],
}

describe("topology view", () => {
  const services = buildServices(
    {
      nodes: [
        { id: "gateway", label: "gateway" },
        { id: "checkout", label: "checkout" },
        { id: "payments", label: "payments" },
      ],
      edges: [
        { source: "gateway", target: "checkout", kind: "http", count: 4 },
        { source: "checkout", target: "payments", kind: "http", count: 2 },
      ],
    },
    signals,
  )

  it("marks a dead dependency critical and leaves an unmatched node unknown", () => {
    const byId = Object.fromEntries(services.map((service) => [service.id, service]))
    expect(byId.payments.status).toBe("critical")
    expect(byId.gateway.status).toBe("healthy")
    expect(byId.checkout.status).toBe("unknown")
    expect(byId.checkout.upstream.map((edge) => edge.other)).toEqual(["gateway"])
    expect(byId.checkout.downstream.map((edge) => edge.other)).toEqual(["payments"])
    expect(byId.gateway.http?.avg_latency_ms).toBe(12)
    expect(byId.checkout.http).toBeNull()
  })

  it("ranks the critical service first and selects the focus from the URL", () => {
    const ranked = rankServices(services)
    expect(ranked[0].id).toBe("payments")
    expect(pickService(ranked, "checkout")?.id).toBe("checkout")
    expect(pickService(ranked, "")?.id).toBe("payments")
    expect(filterServices(ranked, "gate").map((service) => service.id)).toEqual(["gateway"])
  })

  it("separates raw addresses and container ids from named services", () => {
    for (const id of ["10.0.160.97", "2a02:4780:14:3aa7::1", "0aa83922cb55"]) expect(isAddress(id)).toBe(true)
    for (const id of ["gateway", "api.example.com", "orders-response-queue-0aa83922cb55", "beef"]) expect(isAddress(id)).toBe(false)
    const mixed = buildServices({ nodes: [{ id: "gateway", label: "gateway" }, { id: "10.0.0.1", label: "10.0.0.1" }], edges: [] }, signals)
    const { named, addresses } = splitAddresses(mixed)
    expect(named.map((service) => service.id)).toEqual(["gateway"])
    expect(addresses.map((service) => service.id)).toEqual(["10.0.0.1"])
  })
})
