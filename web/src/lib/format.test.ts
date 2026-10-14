import { describe, expect, it } from "vitest"

import i18n from "@/i18n"
import { formatBytes, formatDuration, formatPercent, stripAnsi } from "./format"

describe("format", () => {
  it("strips terminal color codes", () => {
    expect(stripAnsi("\u001b[32minfo\u001b[39m: ready \u001b[1mnow\u001b[0m")).toBe("info: ready now")
    expect(stripAnsi("plain")).toBe("plain")
  })

  it("follows the active locale", async () => {
    await i18n.changeLanguage("en")
    expect(formatPercent(12.345)).toBe("12.3%")
    expect(formatBytes(1536)).toBe("1.5\u00a0KiB")
    expect(formatDuration(1500)).toBe("1.50\u00a0s")
    await i18n.changeLanguage("pt-BR")
    expect(formatPercent(12.345)).toBe("12,3%")
    expect(formatDuration(250)).toBe("250\u00a0ms")
    await i18n.changeLanguage("en")
  })
})
