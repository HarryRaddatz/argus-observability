# Fluxo: painel de gestão

O painel é uma SPA React que consome a API do hub. Componentes visuais exclusivamente shadcn/ui.

## Sequência — carregamento inicial

```mermaid
sequenceDiagram
  participant Browser
  participant Vite as Dev server / nginx
  participant Hub

  Browser->>Vite: GET /
  Vite-->>Browser: index.html + JS
  Browser->>Vite: GET /health
  Vite->>Hub: proxy /health
  Hub-->>Browser: status ok
  Browser->>Vite: GET /api/v1/events
  Vite->>Hub: proxy API
  Hub-->>Browser: timeline JSON
```

## Sequência — métricas no gráfico

```mermaid
sequenceDiagram
  participant UI as Pagina Metricas
  participant Hub
  participant Store

  UI->>Hub: GET /api/v1/query?metric=cpu.usage
  Hub->>Store: aggregate series
  Store-->>Hub: points
  Hub-->>UI: QuerySeries JSON
  UI->>UI: render Chart (shadcn)
```

## Mapa de rotas

```mermaid
flowchart LR
  Shell[AppShell + Sidebar]
  Shell --> Dashboard[Visao geral]
  Shell --> Workloads[Workloads]
  Shell --> Metrics[Metricas]
  Shell --> Logs[Logs]
  Shell --> Events[Eventos]
```

## Camadas da UI

```mermaid
flowchart TB
  Pages[pages/]
  Layout[components/layout]
  UI[components/ui shadcn]
  API[lib/api.ts]
  Pages --> Layout
  Pages --> UI
  Pages --> API
  API --> Hub[Hub REST]
```

## Tema claro / escuro

Variáveis `:root` e `.dark` em `web/src/index.css`. Toggle de tema pode ser adicionado via `next-themes` ou classe no `documentElement` — padrão shadcn.
