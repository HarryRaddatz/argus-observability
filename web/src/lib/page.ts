export type ListPage<T> = {
  entries: T[]
  total: number
  limit: number
  offset: number
  truncated: boolean
}

export const PAGE_SIZE = 25

export function pageIndex(raw: string): number {
  const n = Number(raw)
  if (!Number.isInteger(n) || n < 1) return 1
  return n
}

export function pageSlice<T>(rows: T[], page: number, size: number) {
  const total = rows.length
  const pages = Math.max(1, Math.ceil(total / size))
  const safe = Math.min(Math.max(page, 1), pages)
  const startIndex = (safe - 1) * size
  const slice = rows.slice(startIndex, startIndex + size)
  return {
    page: safe,
    rows: slice,
    total,
    start: total === 0 ? 0 : startIndex + 1,
    end: Math.min(total, startIndex + slice.length),
  }
}

export function asPage<T>(value: unknown): ListPage<T> {
  if (Array.isArray(value)) {
    return { entries: value as T[], total: value.length, limit: value.length, offset: 0, truncated: false }
  }
  if (value && typeof value === "object" && Array.isArray((value as { entries?: unknown }).entries)) {
    const page = value as { entries: T[]; total?: number; limit?: number; offset?: number; truncated?: boolean }
    const entries = page.entries
    return {
      entries,
      total: typeof page.total === "number" ? page.total : entries.length,
      limit: typeof page.limit === "number" ? page.limit : entries.length,
      offset: typeof page.offset === "number" ? page.offset : 0,
      truncated: Boolean(page.truncated),
    }
  }
  return { entries: [], total: 0, limit: 0, offset: 0, truncated: false }
}
