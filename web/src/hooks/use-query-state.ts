import { useCallback } from "react"
import { useSearchParams } from "react-router-dom"

type Patch = Record<string, string | null | undefined>

export function useQueryPatch() {
  const [, setSearchParams] = useSearchParams()
  return useCallback(
    (patch: Patch, defaults: Record<string, string> = {}) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev)
          for (const [key, value] of Object.entries(patch)) {
            if (value === null || value === undefined || value === "" || value === defaults[key]) {
              next.delete(key)
            } else {
              next.set(key, value)
            }
          }
          return next
        },
        { replace: true },
      )
    },
    [setSearchParams],
  )
}

export function useQueryState(key: string, fallback = ""): [string, (value: string) => void] {
  const [searchParams] = useSearchParams()
  const patch = useQueryPatch()
  const value = searchParams.get(key) ?? fallback
  const setValue = useCallback(
    (next: string) => patch({ [key]: next }, { [key]: fallback }),
    [patch, key, fallback],
  )
  return [value, setValue]
}
