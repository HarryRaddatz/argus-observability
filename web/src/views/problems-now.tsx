import { useCallback, useState, type ReactNode } from "react"
import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"

import { LoadError } from "@/components/load-error"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { usePolling } from "@/hooks/use-polling"
import {
  fetchActiveAlerts,
  fetchFleetStatus,
  fetchInsights,
  type ActiveAlert,
  type ContainerFleetStatus,
  type Insight,
} from "@/lib/api"
import { formatDateTime } from "@/lib/format"
import { localizeAlertSummary, localizeAlertTitle } from "@/lib/i18n-catalog"
import { severityLabel, severityVariant } from "@/lib/severity"

type Props = {
  since: string
  group: string
}

export function ProblemsNow({ since, group }: Props) {
  const { t } = useTranslation()
  const [alerts, setAlerts] = useState<ActiveAlert[]>([])
  const [insights, setInsights] = useState<Insight[]>([])
  const [fleet, setFleet] = useState<ContainerFleetStatus[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    Promise.all([fetchActiveAlerts(), fetchInsights(since, group || undefined), fetchFleetStatus()])
      .then(([a, i, f]) => {
        setAlerts(a)
        setInsights(i.insights ?? [])
        setFleet(f.containers ?? [])
        setError(null)
      })
      .catch(() => setError(t("problems.loadError")))
      .finally(() => setLoading(false))
  }, [since, group, t])

  usePolling(load)

  if (loading && alerts.length === 0 && insights.length === 0 && fleet.length === 0) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-14 w-full" />
        ))}
      </div>
    )
  }

  const failed = fleet.filter((c) => c.disposition === "unexpected" || c.disposition === "oom")
  const stopped = fleet.filter((c) => c.disposition === "intentional")

  return (
    <div className="space-y-8">
      {error ? <LoadError message={error} onRetry={load} /> : null}

      <Section title={t("problems.downTitle")}>
        {failed.length === 0 ? (
          <Empty>{t("problems.downEmpty")}</Empty>
        ) : (
          <ul className="divide-y">
            {failed.map((c) => (
              <DownRow key={c.entity_uid} container={c} />
            ))}
          </ul>
        )}
      </Section>

      <Section title={t("problems.intentionalTitle")}>
        {stopped.length === 0 ? (
          <Empty>{t("problems.intentionalEmpty")}</Empty>
        ) : (
          <ul className="divide-y">
            {stopped.map((c) => (
              <DownRow key={c.entity_uid} container={c} />
            ))}
          </ul>
        )}
      </Section>

      <Section title={t("problemsExtra.alerts")}>
        {alerts.length === 0 ? (
          <Empty>{t("problems.noAlerts")}</Empty>
        ) : (
          <ul className="divide-y">
            {alerts.map((a) => (
              <li key={a.rule_id + a.entity_uid} className="flex items-baseline justify-between gap-4 py-3">
                <div className="min-w-0">
                  <p className="text-sm font-medium">{localizeAlertTitle(t, a.rule_id, a.title)}</p>
                  <p className="text-muted-foreground text-sm break-words">
                    {t("problems.since", { container: a.container, when: formatDateTime(a.fired_at) })}
                  </p>
                  {a.summary || a.value ? (
                    <p className="text-muted-foreground text-sm break-words">{localizeAlertSummary(t, a.rule_id, a.value, a.summary)}</p>
                  ) : null}
                  {a.container ? <RowLinks container={a.container} /> : null}
                </div>
                <Badge variant={severityVariant[a.severity] ?? "outline"} className="shrink-0">
                  {severityLabel(t, a.severity)}
                </Badge>
              </li>
            ))}
          </ul>
        )}
      </Section>

      <Section title={t("problemsExtra.insights")}>
        {insights.length === 0 ? (
          <Empty>{t("problems.noInsights")}</Empty>
        ) : (
          <ul className="divide-y">
            {insights.map((ins) => (
              <InsightRow key={ins.id} insight={ins} group={group} />
            ))}
          </ul>
        )}
      </Section>
    </div>
  )
}

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section>
      <h2 className="mb-2 text-sm font-medium">{title}</h2>
      {children}
    </section>
  )
}

function Empty({ children }: { children: ReactNode }) {
  return <p className="text-muted-foreground text-sm">{children}</p>
}

function DownRow({ container }: { container: ContainerFleetStatus }) {
  const { t } = useTranslation()
  const detail =
    container.disposition === "oom"
      ? t("problems.oomDetail")
      : container.disposition === "intentional"
        ? t("problems.intentionalDetail", { code: container.exit_code ?? 0 })
        : t("problems.unexpectedDetail", { code: container.exit_code ?? 0 })
  return (
    <li className="flex items-baseline justify-between gap-4 py-3">
      <div className="min-w-0">
        <p className="text-sm font-medium break-all">{container.container}</p>
        <p className="text-muted-foreground text-sm">{detail}</p>
        <RowLinks container={container.container} />
      </div>
      <Badge variant={container.disposition === "intentional" ? "outline" : "destructive"} className="shrink-0">
        {t(`state.${container.disposition}`)}
      </Badge>
    </li>
  )
}

const rowLink =
  "focus-visible:ring-ring inline-flex items-center rounded-sm underline-offset-4 hover:underline focus-visible:ring-2 focus-visible:outline-none pointer-coarse:min-h-11"

function RowLinks({ container }: { container: string }) {
  const { t } = useTranslation()
  return (
    <p className="mt-1 flex gap-3 text-sm">
      <Link className={rowLink} to={`/logs?container=${encodeURIComponent(container)}`}>
        {t("problems.logsOf")}
      </Link>
      <Link className={rowLink} to={`/metrics?container=${encodeURIComponent(container)}`}>
        {t("problems.metricsOf")}
      </Link>
    </p>
  )
}

function InsightRow({ insight, group }: { insight: Insight; group: string }) {
  const { t } = useTranslation()
  const topic =
    insight.theme === "gc_thrashing"
      ? "gc"
      : insight.theme === "oom_risk" || insight.theme === "memory_pressure"
        ? "memory"
        : insight.theme === "error_spike"
          ? "error"
          : "performance"
  const logsLink = group
    ? `/logs?group=${encodeURIComponent(group)}&topic=${topic}`
    : `/logs?container=${encodeURIComponent(insight.container)}&topic=${topic}`
  const themeKey = `insightTheme.${insight.theme}`
  const theme = t(themeKey)
  const themeLabel = theme === themeKey ? insight.theme : theme

  return (
    <li className="flex items-baseline justify-between gap-4 py-3">
      <div className="min-w-0">
        <p className="text-sm font-medium">{insight.title}</p>
        <p className="text-muted-foreground text-sm">
          <span className="text-foreground">{themeLabel}</span>
          {insight.container ? <span className="break-all">{t("problems.onContainer", { container: insight.container })}</span> : null}
        </p>
        {insight.summary ? <p className="text-muted-foreground text-sm break-words">{insight.summary}</p> : null}
        <p className="mt-1 flex gap-3 text-sm">
          <Link className={rowLink} to={logsLink}>
            {t("problemsExtra.relatedLogs")}
          </Link>
        </p>
      </div>
      <Badge variant={severityVariant[insight.severity] ?? "outline"} className="shrink-0">
        {severityLabel(t, insight.severity)}
      </Badge>
    </li>
  )
}
