import { describe, expect, it } from "vitest"

import { rangeFromLocal, splitRange } from "@/lib/time-range"

describe("rangeFromLocal", () => {
  it("encodes a closed window and rejects one longer than 24 hours", () => {
    const encoded = rangeFromLocal("2026-10-06T14:00", "2026-10-06T15:00")
    expect(encoded).toBeTruthy()
    const parts = splitRange(encoded!)
    expect(parts).not.toBeNull()
    expect(new Date(parts!.end).getTime() - new Date(parts!.start).getTime()).toBe(60 * 60 * 1000)
    expect(rangeFromLocal("2026-10-01T00:00", "2026-10-06T00:00")).toBeNull()
    expect(rangeFromLocal("2026-10-06T15:00", "2026-10-06T14:00")).toBeNull()
  })
})
