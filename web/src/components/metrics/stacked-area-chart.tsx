import { Area, AreaChart, CartesianGrid, Legend, XAxis, YAxis } from "recharts"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart"
import { Skeleton } from "@/components/ui/skeleton"
import type { ContainerSeries } from "@/lib/api"
import { formatTime } from "@/lib/format"
import { containerColor } from "@/lib/chart-colors"

type Props = {
  title: string
  description?: string
  series: ContainerSeries[]
  loading?: boolean
  unit?: string
  transform?: (v: number) => number
  emptyHint?: string
}

export function StackedAreaChart({
  title,
  description,
  series,
  loading,
  unit = "",
  transform,
  emptyHint = "Sem dados no período — tente ampliar o intervalo.",
}: Props) {
  const data = mergeStacked(series, transform)
  const keys = series.map((s) => s.container)

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base">{title}</CardTitle>
        {description ? <CardDescription>{description}</CardDescription> : null}
      </CardHeader>
      <CardContent>
        {loading ? (
          <Skeleton className="h-[280px] w-full" />
        ) : data.length === 0 ? (
          <p className="text-muted-foreground flex h-[280px] items-center justify-center px-4 text-center text-sm">
            {emptyHint}
          </p>
        ) : (
          <ChartContainer config={{}} className="h-[280px] w-full">
            <AreaChart data={data} margin={{ left: 4, right: 8, top: 8, bottom: 0 }}>
              <CartesianGrid vertical={false} strokeDasharray="3 3" />
              <XAxis dataKey="time" tickLine={false} axisLine={false} minTickGap={32} />
              <YAxis tickLine={false} axisLine={false} width={48} tickFormatter={(v) => `${v}${unit}`} />
              <ChartTooltip content={<ChartTooltipContent />} />
              <Legend />
              {keys.map((k, i) => (
                <Area
                  key={k}
                  type="monotone"
                  dataKey={k}
                  name={k}
                  stackId="stack"
                  stroke={containerColor(i)}
                  fill={containerColor(i)}
                  fillOpacity={0.35}
                  strokeWidth={1.5}
                  isAnimationActive={false}
                />
              ))}
            </AreaChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  )
}

function mergeStacked(series: ContainerSeries[], transform?: (v: number) => number) {
  const byTime = new Map<string, Record<string, string | number>>()
  for (const s of series) {
    for (const p of s.points ?? []) {
      const time = formatTime(p.ts)
      const row = byTime.get(time) ?? { time }
      row[s.container] = transform ? transform(p.value) : p.value
      byTime.set(time, row)
    }
  }
  return [...byTime.values()]
}
