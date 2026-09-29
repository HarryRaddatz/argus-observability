import { Navigate, useLocation } from "react-router-dom"

type Props = {
  to: string
  params?: Record<string, string>
}

export function RedirectWithParams({ to, params = {} }: Props) {
  const location = useLocation()
  const search = new URLSearchParams(location.search)
  for (const [key, value] of Object.entries(params)) {
    if (!search.has(key)) search.set(key, value)
  }
  const qs = search.toString()
  return <Navigate to={`${to}${qs ? `?${qs}` : ""}`} replace />
}
