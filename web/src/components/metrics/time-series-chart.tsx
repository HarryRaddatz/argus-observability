import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts"
import { useTranslation } from "react-i18next"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart"
import { Skeleton } from "@/components/ui/skeleton"
import { chartPoints, formatNumber } from "@/lib/format"
import type { SeriesPoint } from "@/lib/api"

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
  const { t } = useTranslation()
  const data = chartPoints(points, transform)
  const config = { value: { label: t("common.value"), color: "var(--chart-2)" } }

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
            {t("chart.empty")}
          </p>
        ) : (
          <ChartContainer
            config={config}
            className={`${height} w-full`}
            role="img"
            aria-label={t("chart.chartLabel", { title, count: data.length })}
          >
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

const SPARK_POINTS = 60

export function Sparkline({ points, label, unit = "" }: { points: SeriesPoint[]; label: string; unit?: string }) {
  const { t } = useTranslation()
  if (points.length === 0) {
    return (
      <div className="bg-muted/40 text-muted-foreground flex h-12 w-full items-center justify-center rounded-md text-xs">
        {t("chart.noData")}
      </div>
    )
  }
  const step = Math.max(1, Math.ceil(points.length / SPARK_POINTS))
  const values: number[] = []
  for (let i = 0; i < points.length; i += step) values.push(points[i].value)
  if ((points.length - 1) % step !== 0) values.push(points[points.length - 1].value)
  const min = Math.min(...values)
  const max = Math.max(...values)
  const span = max - min || 1
  const x = (i: number) => (values.length === 1 ? 50 : (i / (values.length - 1)) * 100)
  const y = (v: number) => 46 - ((v - min) / span) * 42
  const line = values.map((v, i) => `${i === 0 ? "M" : "L"}${x(i).toFixed(2)},${y(v).toFixed(2)}`).join(" ")
  const last = values[values.length - 1]
  const summary = t("chart.sparkSummary", {
    label,
    last: `${formatNumber(last, 1)}${unit}`,
    min: `${formatNumber(min, 1)}${unit}`,
    max: `${formatNumber(max, 1)}${unit}`,
  })
  return (
    <svg viewBox="0 0 100 48" preserveAspectRatio="none" className="text-chart-2 h-12 w-full" role="img" aria-label={summary}>
      <title>{summary}</title>
      <path d={`${line} L100,48 L0,48 Z`} fill="currentColor" fillOpacity={0.2} stroke="none" />
      <path d={line} fill="none" stroke="currentColor" strokeWidth={1.5} vectorEffect="non-scaling-stroke" />
    </svg>
  )
}
