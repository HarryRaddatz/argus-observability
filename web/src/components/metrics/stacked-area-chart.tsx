import { Area, AreaChart, CartesianGrid, Legend, XAxis, YAxis } from "recharts"
import type { TooltipContentProps } from "recharts"
import { useTranslation } from "react-i18next"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { ChartContainer, ChartTooltip } from "@/components/ui/chart"
import { Skeleton } from "@/components/ui/skeleton"
import type { ContainerSeries } from "@/lib/api"
import { containerColor } from "@/lib/chart-colors"
import { formatPercent } from "@/lib/format"
import { mergeStacked } from "@/lib/series"

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
  emptyHint,
}: Props) {
  const { t } = useTranslation()
  const data = mergeStacked(series, transform)
  const keys = series.map((s) => s.container)
  const hint = emptyHint ?? t("chart.emptyHint")

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
            {hint}
          </p>
        ) : (
          <ChartContainer config={{}} className="h-[280px] w-full">
            <AreaChart data={data} margin={{ left: 4, right: 8, top: 8, bottom: 0 }}>
              <CartesianGrid vertical={false} strokeDasharray="3 3" />
              <XAxis dataKey="time" tickLine={false} axisLine={false} minTickGap={32} />
              <YAxis tickLine={false} axisLine={false} width={48} tickFormatter={(v) => `${v}${unit}`} />
              <ChartTooltip content={<StackTooltip unit={unit} />} />
              <Legend />
              {keys.map((k, i) => (
                <Area
                  key={k}
                  type="linear"
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

function StackTooltip({
  active,
  payload,
  label,
  unit,
}: Partial<TooltipContentProps> & { unit: string }) {
  const { t } = useTranslation()
  if (!active || !payload?.length) return null
  const rows = payload
    .filter((item) => typeof item.value === "number" && item.value > 0)
    .sort((a, b) => Number(b.value) - Number(a.value))
  const total = rows.reduce((sum, item) => sum + Number(item.value), 0)
  return (
    <div className="grid min-w-32 items-start gap-1.5 rounded-lg border border-border/50 bg-background px-2.5 py-1.5 text-xs shadow-xl">
      {label ? <div className="font-medium">{label}</div> : null}
      {rows.map((item) => (
        <div key={String(item.dataKey)} className="flex items-center justify-between gap-3">
          <span className="text-muted-foreground flex items-center gap-1.5">
            <span className="h-2.5 w-2.5 shrink-0 rounded-[2px]" style={{ background: item.color }} />
            {item.name}
          </span>
          <span className="font-mono font-medium tabular-nums">
            {unit === "%" ? formatPercent(Number(item.value)) : Number(item.value).toLocaleString()}
          </span>
        </div>
      ))}
      {unit === "%" ? (
        <div className="flex items-center justify-between gap-3 border-t pt-1.5">
          <span className="text-muted-foreground">{t("chart.total")}</span>
          <span className="font-mono font-medium tabular-nums">{formatPercent(total)}</span>
        </div>
      ) : null}
    </div>
  )
}
