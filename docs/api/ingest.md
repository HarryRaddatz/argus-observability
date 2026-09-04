# Ingest API

Rotas POST usadas pelo agent. Auth via `Authorization: Bearer` quando configurado.

## POST `/api/v1/agents/register`

Registra o agent no hub.

**Request**

```json
{
  "agent_id": "argus-agent",
  "host_id": "docker-host",
  "version": "0.1.0",
  "capabilities": ["metrics", "logs", "fleet"]
}
```

**Response** `200`

```json
{
  "status": "registered",
  "agent_id": "argus-agent"
}
```

## POST `/api/v1/agents/heartbeat`

```json
{
  "agent_id": "argus-agent",
  "host_id": "docker-host",
  "ts": "2026-09-04T12:00:00Z"
}
```

## POST `/api/v1/metrics/batch`

```json
{
  "agent_id": "argus-agent",
  "host_id": "docker-host",
  "points": [
    {
      "metric_name": "cpu.usage",
      "ts": "2026-09-04T12:00:00Z",
      "value": 12.5,
      "entity_uid": "container:stack-demo-api-1",
      "labels": { "container": "stack-demo-api-1" }
    }
  ]
}
```

Métricas de infra coletadas pelo agent: `cpu.usage`, `memory.usage`, `memory.limit`, `network.rx`, `network.tx`, `block.read`, `block.write`.

## POST `/api/v1/logs/batch`

```json
{
  "agent_id": "argus-agent",
  "entries": [
    {
      "ts": "2026-09-04T12:00:00Z",
      "level": "info",
      "message": "{\"event\":\"exit\",\"service\":\"demo-api\",\"status\":200,\"durationMs\":120}",
      "entity_uid": "container:stack-demo-api-1",
      "labels": { "container": "stack-demo-api-1" }
    }
  ]
}
```

## POST `/api/v1/fleet/batch`

Snapshot operacional de containers (estado, health, restarts).

## POST `/api/v1/events`

Evento pontual (restart, OOM, rule fired).

## POST `/v1/traces`

Ingest OTLP JSON (spans). Content-Type: `application/json`.

Handler: `internal/hub/otlp.go`

## Códigos de erro

| Código | Situação |
|---|---|
| `401` | Token ausente ou inválido |
| `400` | JSON inválido |
| `500` | Erro de persistência |
