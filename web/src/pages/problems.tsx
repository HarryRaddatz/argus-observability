import { useTranslation } from "react-i18next"

import { GroupSelect } from "@/components/filters/group-select"
import { TimeRangePicker } from "@/components/filters/time-range-picker"
import { PageHeader } from "@/components/layout/page-header"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useQueryPatch, useQueryState } from "@/hooks/use-query-state"
import { EventHistory } from "@/views/event-history"
import { ProblemsNow } from "@/views/problems-now"
import { SLOList } from "@/views/slo-list"

const defaultSince: Record<string, string> = { now: "1h", history: "24h" }

export function ProblemsPage() {
  const { t } = useTranslation()
  const [tab] = useQueryState("tab", "now")
  const [since, setSince] = useQueryState("since", defaultSince[tab] ?? "1h")
  const [group, setGroup] = useQueryState("group", "")
  const patch = useQueryPatch()

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("problems.title")}
        actions={
          tab === "slos" ? null : (
            <>
              {tab === "now" ? <GroupSelect value={group} onChange={setGroup} /> : null}
              <TimeRangePicker value={since} onChange={setSince} />
            </>
          )
        }
      />
      <Tabs value={tab} onValueChange={(v) => patch({ tab: String(v), since: null }, { tab: "now" })}>
        <TabsList aria-label={t("common.section")}>
          <TabsTrigger value="now">{t("problems.now")}</TabsTrigger>
          <TabsTrigger value="slos">{t("problems.slos")}</TabsTrigger>
          <TabsTrigger value="history">{t("problems.history")}</TabsTrigger>
        </TabsList>
      </Tabs>
      {tab === "slos" ? <SLOList /> : tab === "history" ? <EventHistory since={since} /> : <ProblemsNow since={since} group={group} />}
    </div>
  )
}
