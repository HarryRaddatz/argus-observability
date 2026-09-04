import type { ReactNode } from "react"

type Props = {
  title: string
  description?: string
  actions?: ReactNode
  breadcrumb?: string
}

export function PageHeader({ title, description, actions, breadcrumb }: Props) {
  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
      <div className="space-y-1">
        {breadcrumb ? (
          <p className="text-muted-foreground text-xs font-medium uppercase tracking-wide">{breadcrumb}</p>
        ) : null}
        <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
        {description ? <p className="text-muted-foreground max-w-2xl text-sm">{description}</p> : null}
      </div>
      {actions ? <div className="flex shrink-0 flex-wrap items-center gap-2">{actions}</div> : null}
    </div>
  )
}
