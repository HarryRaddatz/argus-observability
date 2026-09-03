# Painel web (Argus)

Interface de gestão construída com [shadcn/ui](https://ui.shadcn.com). Layout e componentes seguem o design system alinhado ao kit Figma oficial: [Figma — shadcn/ui](https://ui.shadcn.com/docs/figma).

## Regra de UI

- **Somente** componentes do registry shadcn (`npx shadcn add …`).
- Tokens, radius e tema via `components.json` e variáveis CSS em `src/index.css`.
- Novas telas: combinar primitives existentes (Card, Sidebar, Table, Chart, Tabs) antes de CSS custom.

## Desenvolvimento

```bash
npm install
npm run dev
```

Proxy Vite encaminha `/api` e `/health` para o hub (default `http://127.0.0.1:8088`).

## Build

```bash
npm run build
```

Artefato em `dist/` — servir via nginx com proxy para o hub.

## Estrutura

```
src/
  components/layout/   shell + sidebar
  components/ui/       shadcn (gerado pelo CLI)
  pages/               rotas do painel
  lib/api.ts           cliente HTTP do hub
```
