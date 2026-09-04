# Fluxo: painel de gestão

SPA React (`web/`) consumindo a API do hub. Componentes shadcn/ui.

## Arquitetura de informação

Navegação agrupada por domínio (`web/src/lib/navigation.ts`):

| Grupo | Rotas |
|---|---|
| Visão | Dashboard (`/`) |
| Infraestrutura | Workloads, Fleet, Grupos |
| Telemetria | Métricas, Explorer, Logs |
| Análise | Insights, Patterns, Topologia, Traces, SLOs |
| Alertas | Eventos |

Header global exibe grupo + título da rota ativa (`app-shell.tsx`).

## Mapa de rotas

```mermaid
flowchart TB
  Shell[AppShell + Sidebar agrupado]
  Shell --> Dashboard
  Shell --> Infra[Workloads / Fleet / Grupos]
  Shell --> Tel[Métricas / Explorer / Logs]
  Shell --> Ana[Insights / Patterns / Topologia / Traces / SLOs]
  Shell --> Events[Eventos]
```

Todas as 13 rotas permanecem estáveis (URLs inalteradas).

## Layout de página

Componente `PageHeader`: título, descrição, breadcrumb opcional, slot de ações (período, toggles).

Páginas principais:

- **Dashboard** — hub com cards por domínio (infra, HTTP, análise)
- **Workloads** — grid com meters/sparklines + tabela; chart CPU empilhado
- **Métricas** — layout 2 colunas em desktop; meters CPU/mem
- **Explorer** — multi-série com toggle média/máx

## Sequência — carregamento

```mermaid
sequenceDiagram
  participant Browser
  participant Web as nginx / Vite
  participant Hub

  Browser->>Web: GET /
  Web-->>Browser: SPA
  Browser->>Hub: GET /health
  Hub-->>Browser: ok
  Browser->>Hub: GET /api/v1/workloads
  Hub-->>Browser: snapshots
```

## Dev

Proxy Vite: `/api` e `/health` → hub (`web/vite.config.ts`).

Build produção: `npm run build` — artefatos servidos pelo container `argus-web`.

## Referências

- Layout: `web/src/components/layout/`
- Métricas UI: `web/src/components/metrics/`
- Mapa produto: [../map.md](../map.md)
