import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart"
import { Skeleton } from "@/components/ui/skeleton"
import { chartPoints } from "@/lib/format"
import type { SeriesPoint } from "@/lib/api"

const chartConfig = {
  value: { label: "Valor", color: "var(--chart-1)" },
}

type Props = {
  title: string
  description?: string
  points: SeriesPoint[]
  loading?: boolean
  unit?: string
  transform?: (v: number) => number
  height?: string
}

export function TimeSeriesChart({
  title,
  description,
  points,
  loading,
  unit = "",
  transform,
  height = "h-[240px]",
}: Props) {
  const data = chartPoints(points, transform)

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base">{title}</CardTitle>
        {description ? <CardDescription>{description}</CardDescription> : null}
      </CardHeader>
      <CardContent>
        {loading ? (
          <Skeleton className={`${height} w-full`} />
        ) : data.length === 0 ? (
          <p className={`text-muted-foreground flex ${height} items-center justify-center text-sm`}>
            Sem dados no período.
          </p>
        ) : (
          <ChartContainer config={chartConfig} className={`${height} w-full`}>
            <AreaChart data={data} margin={{ left: 4, right: 8, top: 8, bottom: 0 }}>
              <CartesianGrid vertical={false} strokeDasharray="3 3" />
              <XAxis dataKey="time" tickLine={false} axisLine={false} minTickGap={32} />
              <YAxis
                tickLine={false}
                axisLine={false}
                width={48}
                tickFormatter={(v) => `${v}${unit}`}
              />
              <ChartTooltip content={<ChartTooltipContent />} />
              <Area
                type="monotone"
                dataKey="value"
                stroke="var(--color-value)"
                fill="var(--color-value)"
                fillOpacity={0.15}
                strokeWidth={2}
                isAnimationActive={false}
              />
            </AreaChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  )
}

export function Sparkline({ points, transform }: { points: SeriesPoint[]; transform?: (v: number) => number }) {
  const data = chartPoints(points, transform)
  if (data.length === 0) {
    return <div className="bg-muted/40 h-12 w-full rounded-md" />
  }
  return (
    <ChartContainer config={chartConfig} className="h-12 w-full">
      <AreaChart data={data} margin={{ top: 4, right: 0, left: 0, bottom: 0 }}>
        <Area
          type="monotone"
          dataKey="value"
          stroke="var(--color-value)"
          fill="var(--color-value)"
          fillOpacity={0.2}
          strokeWidth={1.5}
          isAnimationActive={false}
        />
      </AreaChart>
    </ChartContainer>
  )
}
