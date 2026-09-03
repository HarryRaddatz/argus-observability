import { Area, AreaChart, CartesianGrid, Legend, Line, LineChart, XAxis, YAxis } from "recharts"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart"
import { Skeleton } from "@/components/ui/skeleton"
import { formatTime } from "@/lib/format"
import type { ContainerSeries } from "@/lib/api"

const COLORS = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
]

type Props = {
  title: string
  description?: string
  series: ContainerSeries[]
  loading?: boolean
  unit?: string
  transform?: (v: number) => number
  chartType?: "area" | "line"
}

export function MultiSeriesChart({
  title,
  description,
  series,
  loading,
  unit = "",
  transform,
  chartType = "area",
}: Props) {
  const data = mergeSeries(series, transform)
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
          <p className="text-muted-foreground flex h-[280px] items-center justify-center text-sm">
            Sem dados no período.
          </p>
        ) : chartType === "line" ? (
          <ChartContainer config={{}} className="h-[280px] w-full">
            <LineChart data={data} margin={{ left: 4, right: 8, top: 8, bottom: 0 }}>
              <CartesianGrid vertical={false} strokeDasharray="3 3" />
              <XAxis dataKey="time" tickLine={false} axisLine={false} minTickGap={32} />
              <YAxis tickLine={false} axisLine={false} width={48} tickFormatter={(v) => `${v}${unit}`} />
              <ChartTooltip content={<ChartTooltipContent />} />
              <Legend />
              {keys.map((k, i) => (
                <Line
                  key={k}
                  type="monotone"
                  dataKey={k}
                  name={k}
                  stroke={COLORS[i % COLORS.length]}
                  strokeWidth={2}
                  dot={false}
                  isAnimationActive={false}
                />
              ))}
            </LineChart>
          </ChartContainer>
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
                  stroke={COLORS[i % COLORS.length]}
                  fill={COLORS[i % COLORS.length]}
                  fillOpacity={0.12}
                  strokeWidth={2}
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

function mergeSeries(series: ContainerSeries[], transform?: (v: number) => number) {
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
