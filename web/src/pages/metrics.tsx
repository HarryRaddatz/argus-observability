import { useEffect, useMemo, useState } from "react"
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart"
import { Skeleton } from "@/components/ui/skeleton"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { queryMetrics } from "@/lib/api"

const chartConfig = {
  value: { label: "Valor", color: "var(--chart-1)" },
}

export function MetricsPage() {
  const [cpu, setCpu] = useState<{ time: string; value: number }[]>([])
  const [mem, setMem] = useState<{ time: string; value: number }[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([
      queryMetrics("cpu.usage", "1h"),
      queryMetrics("memory.usage", "1h"),
    ])
      .then(([cpuSeries, memSeries]) => {
        setCpu(
          cpuSeries.points.map((p) => ({
            time: new Date(p.ts).toLocaleTimeString(),
            value: Math.round(p.value * 10) / 10,
          })),
        )
        setMem(
          memSeries.points.map((p) => ({
            time: new Date(p.ts).toLocaleTimeString(),
            value: Math.round(p.value / 1024 / 1024),
          })),
        )
      })
      .catch(() => {
        setCpu([])
        setMem([])
      })
      .finally(() => setLoading(false))
  }, [])

  const hasData = useMemo(() => cpu.length > 0 || mem.length > 0, [cpu, mem])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Métricas</h1>
        <p className="text-muted-foreground text-sm">
          Séries agregadas recebidas dos agents.
        </p>
      </div>
      <Tabs defaultValue="cpu">
        <TabsList>
          <TabsTrigger value="cpu">CPU</TabsTrigger>
          <TabsTrigger value="memory">Memória</TabsTrigger>
        </TabsList>
        <TabsContent value="cpu" className="mt-4">
          <MetricChart
            title="cpu.usage"
            description="Percentual médio por intervalo"
            data={cpu}
            loading={loading}
            empty={!hasData}
            unit="%"
          />
        </TabsContent>
        <TabsContent value="memory" className="mt-4">
          <MetricChart
            title="memory.usage"
            description="Uso em MiB"
            data={mem}
            loading={loading}
            empty={!hasData}
            unit="MiB"
          />
        </TabsContent>
      </Tabs>
    </div>
  )
}

function MetricChart({
  title,
  description,
  data,
  loading,
  empty,
  unit,
}: {
  title: string
  description: string
  data: { time: string; value: number }[]
  loading: boolean
  empty: boolean
  unit: string
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent>
        {loading ? (
          <Skeleton className="h-[280px] w-full" />
        ) : empty || data.length === 0 ? (
          <p className="text-muted-foreground flex h-[280px] items-center justify-center text-sm">
            Sem pontos ainda. Confirme que o agent está enviando métricas.
          </p>
        ) : (
          <ChartContainer config={chartConfig} className="h-[280px] w-full">
            <AreaChart data={data} margin={{ left: 8, right: 8 }}>
              <CartesianGrid vertical={false} />
              <XAxis dataKey="time" tickLine={false} axisLine={false} minTickGap={24} />
              <YAxis tickLine={false} axisLine={false} width={40} unit={unit === "%" ? "%" : undefined} />
              <ChartTooltip content={<ChartTooltipContent />} />
              <Area
                type="monotone"
                dataKey="value"
                stroke="var(--color-value)"
                fill="var(--color-value)"
                fillOpacity={0.2}
              />
            </AreaChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  )
}
