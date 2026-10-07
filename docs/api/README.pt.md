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

Rotas de ingest e de consulta exigem o header quando `ARGUS_AGENT_TOKEN` está configurado:

```
Authorization: Bearer <token>
```

`/health` continua aberto. Token vazio desliga a checagem. O proxy do painel adiciona o header; o browser não vê o token.

## Base URL

```
http://localhost:8080
```
