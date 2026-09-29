import type { ReactNode } from "react"
import { Link } from "react-router-dom"

import { Card, CardContent } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { textTone, type MeterTone } from "@/lib/meter"
import { cn } from "@/lib/utils"

type Props = {
  label: string
  value: ReactNode
  tone?: MeterTone
  hint?: string
  to?: string
  loading?: boolean
}

export function StatCard({ label, value, tone = "default", hint, to, loading }: Props) {
  const body = (
    <Card className={cn("h-full", to && "hover:bg-muted/40 transition-colors")}>
      <CardContent className="space-y-1 pt-6">
        {loading ? (
          <Skeleton className="h-8 w-16" />
        ) : (
          <p className={cn("text-2xl font-semibold tabular-nums", textTone[tone])}>{value}</p>
        )}
        <p className="text-muted-foreground text-sm">{label}</p>
        {hint ? <p className="text-muted-foreground text-xs">{hint}</p> : null}
      </CardContent>
    </Card>
  )
  if (!to) return body
  return (
    <Link to={to} className="focus-visible:ring-ring block rounded-xl focus-visible:ring-2 focus-visible:outline-none">
      {body}
    </Link>
  )
}
