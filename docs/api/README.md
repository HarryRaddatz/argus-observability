# API reference

HTTP documentation for the Argus hub.

## Index

| Doc | Scope |
|---|---|
| [configuration.md](configuration.md) | Environment variables |
| [ingest.md](ingest.md) | Agents, metrics, logs, fleet, events, OTLP |
| [query.md](query.md) | Workloads, metrics, logs, events, groups |
| [observability.md](observability.md) | Insights, patterns, topology, traces, SLOs, alerts |

Shared contracts: `internal/model/types.go`

## Authentication

Ingest routes require a header when `ARGUS_AGENT_TOKEN` is set:

```
Authorization: Bearer <token>
```

GET queries are open by default (no auth).

## Base URL

```
http://localhost:8080
```
