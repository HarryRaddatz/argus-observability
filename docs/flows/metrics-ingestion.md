# Flow: metrics ingest

The agent queries the runtime on an interval, normalizes generic points, and sends them to the hub in batches.

## Sequence — collect and persist

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

## Point format

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

The hub stores points as they arrive and derives `memory.usage_pct` when the batch includes usage and limit. There is no rollup.

## Query

```mermaid
sequenceDiagram
  participant Client
  participant Hub
  participant Store

  Client->>Hub: GET /api/v1/metrics/series?metric=cpu.usage&since=1h
  Hub->>Store: QueryMetricSeries
  Store-->>Hub: series by container
  Hub-->>Client: metric_name, series
```

`GET /api/v1/query?metric=cpu.usage&since=1h` returns the flattened series (`metric_name`, `points`). There is no `host` filter.

## Agent metrics

| metric_name | Description |
|---|---|
| `cpu.usage` | container CPU % |
| `memory.usage` | bytes |
| `memory.limit` | cgroup bytes |
| `memory.usage_pct` | derived on the hub |
| `network.rx` / `network.tx` | bytes/s |
| `block.read` / `block.write` | bytes/s |
