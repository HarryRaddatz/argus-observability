# Query API

Rotas GET de consulta. Parâmetro `since` aceita duração Go (`30m`, `1h`, `24h`).

## GET `/health`

```json
{ "status": "ok" }
```

## GET `/api/v1/workloads`

Último snapshot por container monitorado.

**Exemplo:** `GET /api/v1/workloads?since=30m`

```json
[
  {
    "container": "stack-demo-api-1",
    "entity_uid": "container:stack-demo-api-1",
    "cpu_usage": 4.2,
    "memory_usage": 134217728,
    "memory_limit": 536870912,
    "updated_at": "2026-09-04T12:00:00Z"
  }
]
```

## GET `/api/v1/metrics/series`

Série temporal agregada.

**Exemplo:** `GET /api/v1/metrics/series?metric=cpu.usage&since=1h&container=stack-demo-api-1`

```json
{
  "metric_name": "cpu.usage",
  "series": [
    {
      "container": "stack-demo-api-1",
      "entity_uid": "container:stack-demo-api-1",
      "points": [
        { "ts": "2026-09-04T11:00:00Z", "value": 3.1 },
        { "ts": "2026-09-04T11:15:00Z", "value": 5.8 }
      ]
    }
  ]
}
```

Filtros opcionais: `container`, `group` (ID de workload group).

## GET `/api/v1/metrics/catalog`

Lista métricas disponíveis no explorer.

## GET `/api/v1/metrics/http/summary`

Resumo HTTP derivado de logs por serviço (`service` label).

## GET `/api/v1/logs/search`

**Exemplo:** `GET /api/v1/logs/search?since=1h&container=stack-demo-api-1&q=error`

```json
{
  "entries": [
    {
      "ts": "2026-09-04T12:00:00Z",
      "level": "error",
      "message": "connection refused",
      "entity_uid": "container:stack-demo-api-1",
      "labels": { "container": "stack-demo-api-1" }
    }
  ]
}
```

## GET `/api/v1/events`

Timeline de eventos. Query: `since`, `entity_uid`.

## GET `/api/v1/fleet/status`

Estado operacional agregado + lista de containers.

## Workload groups

| Método | Rota |
|---|---|
| GET | `/api/v1/workload-groups` |
| GET | `/api/v1/workload-groups/discover` |
| POST | `/api/v1/workload-groups` |
| GET | `/api/v1/workload-groups/{id}` |
| GET | `/api/v1/workload-groups/{id}/summary?since=1h` |
| PUT | `/api/v1/workload-groups/{id}` |
| DELETE | `/api/v1/workload-groups/{id}` |
