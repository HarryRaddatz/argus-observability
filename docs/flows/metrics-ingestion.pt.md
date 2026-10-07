# Fluxo: ingestão de métricas

O agent consulta o runtime a cada intervalo, normaliza pontos genéricos e envia em lote ao hub.

## Sequência — coleta e persistência

```mermaid
sequenceDiagram
  participant Agent
  participant Spool
  participant Hub
  participant Store

  loop tick interval
    Agent->>Agent: batch_id
    Agent->>Hub: POST /api/v1/metrics/batch
    Note over Agent,Hub: Bearer e X-Argus-Batch-Id
    alt 2xx e id novo
      Hub->>Store: metric_points e ingest_batches
      Hub-->>Agent: 202
    else batch_id já gravado
      Hub-->>Agent: 202 sem linhas novas
    else 429, 5xx ou rede
      Agent->>Agent: até cinco tentativas
      Agent->>Spool: guarda batch_id e corpo
    else 401, 403 ou outro 4xx
      Hub-->>Agent: falha, sem spool
    end
  end
```

No tick seguinte o agent reenvia o spool, do mais antigo para o mais novo, antes do lote novo. Cada replay tem o próprio timeout. Registro e heartbeat não entram no spool.

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
  Note over Client,Hub: Bearer quando ARGUS_AGENT_TOKEN está definido
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
