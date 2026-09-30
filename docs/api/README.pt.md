# API reference

Documentação HTTP do hub Argus.

## Índice

| Doc | Escopo |
|---|---|
| [configuration.md](configuration.md) | Variáveis de ambiente |
| [ingest.md](ingest.md) | Agents, metrics, logs, fleet, events, OTLP |
| [query.md](query.md) | Workloads, métricas, logs, eventos, grupos |
| [observability.md](observability.md) | Insights, patterns, topology, traces, SLOs, alertas |

Contratos compartilhados: `internal/model/types.go`

## Autenticação

Rotas de ingest exigem header quando `ARGUS_AGENT_TOKEN` está configurado:

```
Authorization: Bearer <token>
```

Consultas GET são abertas por default (sem auth).

## Base URL

```
http://localhost:8080
```
