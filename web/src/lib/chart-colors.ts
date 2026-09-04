const PALETTE = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
  "hsl(262 83% 58%)",
  "hsl(173 58% 39%)",
  "hsl(43 74% 49%)",
  "hsl(0 72% 51%)",
  "hsl(199 89% 48%)",
]

export function containerColor(index: number): string {
  return PALETTE[index % PALETTE.length]
}
