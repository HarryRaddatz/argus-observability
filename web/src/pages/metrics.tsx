import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { GroupSelect } from "@/components/filters/group-select"
import { TimeRangePicker } from "@/components/filters/time-range-picker"
import { PageHeader } from "@/components/layout/page-header"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useQueryPatch, useQueryState } from "@/hooks/use-query-state"
import { getWorkloadGroupSummary, listWorkloads, type WorkloadSnapshot } from "@/lib/api"
import { ContainerMetrics } from "@/views/container-metrics"
import { MetricCompare } from "@/views/metric-compare"

export function MetricsPage() {
  const { t } = useTranslation()
  const [mode] = useQueryState("mode", "container")
  const [since, setSince] = useQueryState("since", "1h")
  const [group, setGroup] = useQueryState("group", "")
  const patch = useQueryPatch()
  const [workloads, setWorkloads] = useState<WorkloadSnapshot[]>([])
  const [members, setMembers] = useState<Set<string> | null>(null)

  useEffect(() => {
    listWorkloads("1h").then(setWorkloads).catch(() => setWorkloads([]))
  }, [])

  useEffect(() => {
    if (!group) {
      setMembers(null)
      return
    }
    getWorkloadGroupSummary(group, "1h")
      .then((s) => setMembers(new Set(s.members.map((m) => m.container))))
      .catch(() => setMembers(new Set()))
  }, [group])

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("metrics.title")}
        actions={
          <>
            <GroupSelect value={group} onChange={setGroup} />
            <TimeRangePicker value={since} onChange={setSince} />
          </>
        }
      />
      <Tabs value={mode} onValueChange={(v) => patch({ mode: String(v) }, { mode: "container" })}>
        <TabsList aria-label={t("common.mode")}>
          <TabsTrigger value="container">{t("metrics.byContainer")}</TabsTrigger>
          <TabsTrigger value="compare">{t("metrics.compare")}</TabsTrigger>
        </TabsList>
      </Tabs>
      {mode === "compare" ? (
        <MetricCompare since={since} group={group} workloads={workloads} />
      ) : (
        <ContainerMetrics since={since} workloads={workloads} members={members} />
      )}
    </div>
  )
}
