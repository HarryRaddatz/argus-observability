import { describe, expect, it } from "vitest"

import { parseStatMode, summarize } from "./stats"

describe("summarize", () => {
  const values = Array.from({ length: 101 }, (_, i) => i)

  it("returns average, peak, and interpolated percentiles", () => {
    expect(summarize(values, "avg")).toBe(50)
    expect(summarize(values, "max")).toBe(100)
    expect(summarize(values, "p95")).toBeCloseTo(95)
    expect(summarize(values, "p90")).toBeCloseTo(90)
    expect(summarize(values, "p85")).toBeCloseTo(85)
    expect(summarize([10, 20], "p95")).toBeCloseTo(19.5)
  })

  it("ignores input order and handles empty or single values", () => {
    expect(summarize([5, 1, 9, 3], "max")).toBe(9)
    expect(summarize([7], "p90")).toBe(7)
    expect(summarize([], "p95")).toBeNull()
  })

  it("falls back to average for unknown modes", () => {
    expect(parseStatMode("p99")).toBe("avg")
    expect(parseStatMode("p90")).toBe("p90")
  })
})
