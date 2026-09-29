import { useEffect } from "react"

export const REFRESH_MS = 30_000

export function usePolling(load: () => void, intervalMs: number = REFRESH_MS) {
  useEffect(() => {
    load()
    const t = setInterval(load, intervalMs)
    return () => clearInterval(t)
  }, [load, intervalMs])
}
