# Changelog

Formato baseado em [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased]

### Added

- Workflow de release: imagens GHCR + GitHub Release a partir de tags semver
- Compose `examples/compose-minimal/docker-compose.published.yml` (pull GHCR)

### Fixed

- CI: `InferServiceFromContainer` para serviços compose com hífen (`demo-api`)
- Gate anti-leak no workflow (#25)

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

[0.1.0]: https://github.com/HarryRaddatz/argus-observability/releases/tag/v0.1.0
