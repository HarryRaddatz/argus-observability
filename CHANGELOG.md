# Changelog

Formato baseado em [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased]

### Added

- `GET /api/v1/traces`: traces recentes de spans OTLP e de logs com `trace_id`, com filtro por serviço e período (#48)
- `GET /api/v1/logs/patterns` aceita `container`, `group` e `q` (#48)
- Painel: lista de traces por serviço e período, link "Pular para o conteúdo" e menu num landmark `nav` (#48)

### Changed

- Referência da API e fluxos alinhados aos handlers: variáveis de `.env.example`, esquemas de topologia, alertas e SLO, e ingest de fleet, eventos e OTLP (#50)
- Painel com 7 destinos em vez de 13: Visão geral, Problemas, Containers, Métricas, Logs, Traces e Topologia. Workloads, Fleet e Grupos viram Containers; Explorer vira o modo Comparar de Métricas; Patterns vira o modo Padrões de Logs; Eventos, Insights e SLOs viram abas de Problemas. As rotas antigas redirecionam mantendo os parâmetros (#48)
- Visão geral mostra só o que pede ação: alertas ativos, SLOs em risco, containers com problema e serviços HTTP com erro (#48)
- Modo, aba, grade ou tabela, tipo de gráfico e containers comparados ficam na URL (#48)
- Um único título por tela, igual ao nome no menu; sem rótulos em caixa alta nem metadados unidos por `·` (#48)
- CI roda `go`, `web` e o build de cada imagem só quando hub, agent ou web mudam no diff; tags e mudanças no próprio CI rodam tudo (#45)
- `CONTRIBUTING.md`: comandos locais do painel iguais aos do CI (`npm run lint`) e versão do Node
- READMEs das imagens no Docker Hub: tags `vX.Y.Z` e `X`, imagem base, espelho no GHCR e variáveis de purge (hub) e de intervalo/filtro (agent)

### Fixed

- Docs de configuração: `ARGUS_LOG_INTERVAL`, `ARGUS_FLEET_INTERVAL` e `ARGUS_NAME_PREFIX` do agent não estavam documentadas
- Link de SLO para Métricas enviava `service`, que a tela ignorava; agora abre a comparação dos containers do serviço (#48)
- Comparação de métricas HTTP não encontrava as séries, que chegam com o nome do serviço no lugar do container (#48)

## [0.2.1] - 2026-09-29

### Changed

- Go 1.27, Node 24 (LTS), Alpine 3.24, nginx 1.30 (stable) e Python 3.14 nos Dockerfiles e workflows (#41)
- Runners fixados em `ubuntu-26.04` (LTS) e actions atualizadas para as majors atuais (#41)
- `modernc.org/sqlite` 1.60 e demais dependencias Go; dependencias do web nas versoes minor/patch atuais (#43)

### Security

- `undici` atualizado no web, corrigindo vulnerabilidade moderada apontada pelo `npm audit` (#43)

## [0.2.0] - 2026-09-29

### Added

- Workflow de release: imagens GHCR + GitHub Release a partir de tags semver
- Compose `examples/compose-minimal/docker-compose.published.yml` (pull GHCR)
- Runbook de atualização de stack em `docs/deploy.md` (pull GHCR ou build local, checklist pós-deploy, rollback)
- Checklist semver e de publicação no `CONTRIBUTING.md`
- Site de documentação no GitHub Pages (MkDocs), publicado a cada merge em `main` e a cada release
- CI com job de lint do web e build dos três Dockerfiles
- Imagens no Docker Hub (`pseudohuery/argus-{hub,agent,web}`) além do GHCR, com README e descrição sincronizados
- Imagens multi-arch `linux/amd64` e `linux/arm64`, com SBOM, provenance e labels OCI; tags `X.Y` e `X` (a partir de `1.0`)
- `ARGUS_REGISTRY` no compose publicado para escolher entre GHCR e Docker Hub

### Changed

- CI em jobs paralelos (`anti-leak`, `go`, `web`, `docker`) com agregador `test` e cancelamento de execuções antigas do mesmo PR
- Release valida tag, commit em `main` e seção do CHANGELOG antes de publicar; reutiliza o CI; imagens em paralelo com cache; tags `-rc` viram pre-release sem mover `latest`
- CI usa Node 22, a mesma versão do `web/Dockerfile`

### Fixed

- CI: `InferServiceFromContainer` para serviços compose com hífen (`demo-api`)
- Gate anti-leak no workflow (#25)
- Release: caminho do Dockerfile do web (`./web/Dockerfile`) no build da imagem
- Docs de configuração: defaults reais de `ARGUS_STORE_PATH`, `ARGUS_HUB_URL`, `ARGUS_AGENT_ID` e `ARGUS_HOST_ID`

## [0.1.0] - 2026-09-04

### Added

- Hub (`argus-hub`) com API REST, ingest de métricas/logs/eventos/fleet e OTLP traces
- Agent Docker com coleta de CPU, memória, rede e block I/O
- Painel web (React + shadcn/ui): dashboard, workloads, métricas, explorer, logs, insights, patterns, topologia, traces, SLOs, eventos
- Sidebar agrupada por domínio (Visão, Infraestrutura, Telemetria, Análise, Alertas)
- Gráficos de infraestrutura: meters, sparklines, charts empilhados
- Documentação pública: README, CONTRIBUTING, API reference, examples
- CI GitHub Actions (Go test + build web)
- Licença MIT

### Changed

- Seeds e exemplos genéricos (`demo-api`) — sem referências a infra privada
- Portas default do compose: hub `8080`, painel `3000`

[0.2.1]: https://github.com/HarryRaddatz/argus-observability/releases/tag/v0.2.1
[0.2.0]: https://github.com/HarryRaddatz/argus-observability/releases/tag/v0.2.0
[0.1.0]: https://github.com/HarryRaddatz/argus-observability/releases/tag/v0.1.0
