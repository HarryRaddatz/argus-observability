export function WorkloadsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Workloads</h1>
        <p className="text-muted-foreground text-sm">
          Inventário de entidades monitoradas (host, container, pod).
        </p>
      </div>
      <p className="text-muted-foreground text-sm">
        A listagem dinâmica será alimentada pela API de entidades. Métricas e logs já
        carregam <code className="text-xs">entity_uid</code> por workload.
      </p>
    </div>
  )
}
