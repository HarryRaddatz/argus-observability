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

O hub grava os pontos como chegam e deriva `memory.usage_pct` quando o lote traz uso e limite. Não há rollup.

## Consulta

```mermaid
sequenceDiagram
  participant Client
  participant Hub
  participant Store

  Client->>Hub: GET /api/v1/metrics/series?metric=cpu.usage&since=1h
  Hub->>Store: QueryMetricSeries
  Store-->>Hub: series por container
  Hub-->>Client: metric_name, series
```

`GET /api/v1/query?metric=cpu.usage&since=1h` devolve a série achatada (`metric_name`, `points`). Não há filtro `host`.

## Métricas do agent

| metric_name | Descrição |
|---|---|
| `cpu.usage` | % de CPU do container |
| `memory.usage` | bytes |
| `memory.limit` | bytes do cgroup |
| `memory.usage_pct` | derivada no hub |
| `network.rx` / `network.tx` | bytes/s |
| `block.read` / `block.write` | bytes/s |
