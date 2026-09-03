# Fluxo: ingestão de métricas

O agent consulta o runtime a cada intervalo, normaliza pontos genéricos e envia em lote ao hub.

## Sequência — coleta e persistência

```mermaid
sequenceDiagram
  participant Runtime as Runtime API
  participant Agent
  participant Hub
  participant Store

  loop tick interval
    Agent->>Runtime: stats host + workloads
    Runtime-->>Agent: cpu, mem, net, disk
    Agent->>Agent: attach entity_uid + labels
    Agent->>Hub: POST /api/v1/metrics/batch
    Hub->>Store: insert metric_points
    Hub-->>Agent: 202 accepted
  end
```

## Formato de ponto

```json
{
  "metric_name": "cpu.usage",
  "ts": "2026-09-03T13:00:00Z",
  "value": 42.5,
  "entity_uid": "docker:host-01:api",
  "labels": {
    "host": "host-01",
    "runtime": "docker",
    "container": "api"
  }
}
```

## Downsample

```mermaid
flowchart LR
  Raw[Pontos raw 15s] --> Rollup[Rollup 1m no hub]
  Rollup --> Store[(Store)]
  Raw --> Store
```

Retenção configurável: raw curto, rollup longo.

## Consulta (UI / MCP)

```mermaid
sequenceDiagram
  participant Client
  participant Hub
  participant Store

  Client->>Hub: GET /api/v1/query?metric=cpu.usage&host=host-01
  Hub->>Store: aggregate by interval
  Store-->>Hub: series
  Hub-->>Client: JSON timeseries
```

## Métricas host (v1)

| metric_name | Descrição |
|---|---|
| `cpu.usage` | % uso CPU |
| `memory.usage` | bytes ou % |
| `memory.limit` | limite cgroup |
| `disk.usage` | bytes usados |
| `network.rx` / `network.tx` | bytes/s |
| `load.1m` | load average |

## Métricas workload

| metric_name | Descrição |
|---|---|
| `cpu.usage` | % do container/pod |
| `memory.usage` | bytes |
| `memory.limit` | bytes |
