import { useCallback, useState } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"

import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { usePolling } from "@/hooks/use-polling"
import { listEvents, type EventRow } from "@/lib/api"
import { formatDateTime } from "@/lib/format"
import { severityLabel, severityVariant } from "@/lib/severity"

function formatPayload(payload?: Record<string, unknown>) {
  if (!payload || Object.keys(payload).length === 0) return "—"
  return Object.entries(payload)
    .slice(0, 3)
    .map(([k, v]) => `${k}=${String(v)}`)
    .join(", ")
}

export function EventHistory({ since }: { since: string }) {
  const { t } = useTranslation()
  const [rows, setRows] = useState<EventRow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    listEvents(undefined, since)
      .then((ev) => {
        setRows(ev)
        setError(null)
      })
      .catch((e) => setError(e instanceof Error ? e.message : t("events.loadError")))
      .finally(() => setLoading(false))
  }, [since, t])

  usePolling(load)

  return (
    <>
      {error ? <p className="text-destructive text-sm">{error}</p> : null}
      <ScrollArea className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("events.time")}</TableHead>
              <TableHead>{t("events.type")}</TableHead>
              <TableHead>{t("events.severity")}</TableHead>
              <TableHead>{t("common.container")}</TableHead>
              <TableHead>{t("events.source")}</TableHead>
              <TableHead>{t("events.payload")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6}>
                  <Skeleton className="h-8 w-full" />
                </TableCell>
              </TableRow>
            ) : rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-muted-foreground text-center">
                  {t("events.empty")}
                </TableCell>
              </TableRow>
            ) : (
              rows.map((e) => {
                const container = e.entity_uid.split(":").pop() ?? e.entity_uid
                return (
                  <TableRow key={e.id}>
                    <TableCell className="whitespace-nowrap font-mono text-xs">{formatDateTime(e.ts)}</TableCell>
                    <TableCell className="font-medium">{e.type}</TableCell>
                    <TableCell>
                      <Badge variant={severityVariant[e.severity] ?? "outline"}>{severityLabel(t, e.severity)}</Badge>
                    </TableCell>
                    <TableCell className="max-w-[180px] truncate font-mono text-xs">
                      <Link to={`/logs?container=${encodeURIComponent(container)}`} className="text-primary hover:underline">
                        {container}
                      </Link>
                    </TableCell>
                    <TableCell className="text-muted-foreground text-sm">{e.source}</TableCell>
                    <TableCell className="text-muted-foreground max-w-[240px] truncate text-xs">{formatPayload(e.payload)}</TableCell>
                  </TableRow>
                )
              })
            )}
          </TableBody>
        </Table>
      </ScrollArea>
    </>
  )
}
