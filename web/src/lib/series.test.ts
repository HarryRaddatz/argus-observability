import { describe, expect, it } from "vitest"

import { mergeStacked } from "@/lib/series"

describe("mergeStacked", () => {
  it("averages samples in the same minute and fills a gap with zero", () => {
    const rows = mergeStacked([
      {
        container: "a",
        points: [
          { ts: "2026-10-06T18:01:10Z", value: 10 },
          { ts: "2026-10-06T18:01:40Z", value: 30 },
          { ts: "2026-10-06T18:02:10Z", value: 4 },
        ],
      },
      {
        container: "b",
        points: [{ ts: "2026-10-06T18:02:10Z", value: 6 }],
      },
    ])
    expect(rows).toHaveLength(2)
    expect(rows[0].a).toBe(20)
    expect(rows[0].b).toBe(0)
    expect(rows[1].a).toBe(4)
    expect(rows[1].b).toBe(6)
  })
})
