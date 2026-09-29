# Changelog

Formato baseado em [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased]

### Changed

- Go 1.27, Node 24 (LTS), Alpine 3.24, nginx 1.30 (stable) e Python 3.14 nos Dockerfiles e workflows (#41)
- Runners fixados em `ubuntu-26.04` (LTS) e actions atualizadas para as majors atuais (#41)

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

[0.2.0]: https://github.com/HarryRaddatz/argus-observability/releases/tag/v0.2.0
[0.1.0]: https://github.com/HarryRaddatz/argus-observability/releases/tag/v0.1.0
