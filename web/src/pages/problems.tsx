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
  const [tab] = useQueryState("tab", "now")
  const [since, setSince] = useQueryState("since", defaultSince[tab] ?? "1h")
  const [group, setGroup] = useQueryState("group", "")
  const patch = useQueryPatch()

  return (
    <div className="space-y-6">
      <PageHeader
        title="Problemas"
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
        <TabsList aria-label="Seção">
          <TabsTrigger value="now">Agora</TabsTrigger>
          <TabsTrigger value="slos">SLOs</TabsTrigger>
          <TabsTrigger value="history">Histórico de eventos</TabsTrigger>
        </TabsList>
      </Tabs>
      {tab === "slos" ? <SLOList /> : tab === "history" ? <EventHistory since={since} /> : <ProblemsNow since={since} group={group} />}
    </div>
  )
}
