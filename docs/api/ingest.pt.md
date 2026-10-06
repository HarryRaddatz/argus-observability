# Ingest API

Rotas POST usadas pelo agent. Auth via `Authorization: Bearer` quando configurado.

## POST `/api/v1/agents/register`

Registra o agent no hub.

**Request** (`agent_id` e `host_id` obrigatórios; `runtime` vazio vira `docker`)

```json
{
  "agent_id": "agent-1",
  "host_id": "docker-host",
  "runtime": "docker",
  "labels": { "host": "docker-host" }
}
```

**Response** `200`

```json
{
  "session_id": "4f1c0b2e-7a11-4c3d-9e2a-6b8d0f1a2c33",
  "interval_seconds": 30
}
```

`interval_seconds` é o intervalo de heartbeat do hub (default 30). Não é `ARGUS_COLLECT_INTERVAL`.

## POST `/api/v1/agents/heartbeat`

O handler lê só `agent_id`.

```json
{ "agent_id": "agent-1" }
```

**Response** `200`

```json
{ "status": "ok" }
```

## POST `/api/v1/metrics/batch`

Corpo: array de `MetricPoint`. Array vazio responde `202` sem persistir.

```json
[
  {
    "metric_name": "cpu.usage",
    "ts": "2026-09-04T12:00:00Z",
    "value": 12.5,
    "entity_uid": "container:stack-demo-api-1",
    "labels": { "container": "stack-demo-api-1" }
  }
]
```

**Response** `202` sem corpo.

Métricas de infra coletadas pelo agent: `cpu.usage`, `memory.usage`, `memory.limit`, `network.rx`, `network.tx`, `block.read`, `block.write`.

## POST `/api/v1/logs/batch`

Corpo: array de `LogEntry`. Array vazio responde `202` sem persistir.

```json
[
  {
    "ts": "2026-09-04T12:00:00Z",
    "level": "info",
    "message": "{\"event\":\"exit\",\"service\":\"demo-api\",\"status\":200,\"durationMs\":120}",
    "entity_uid": "container:stack-demo-api-1",
    "labels": { "container": "stack-demo-api-1" }
  }
]
```

**Response** `202` sem corpo. O hub preenche `fields` na ingestão; o agent não precisa enviá-los.

## POST `/api/v1/fleet/batch`

Array de `ContainerFleetStatus`. Array vazio responde `202` sem gravar.

```json
[
  {
    "container": "stack-demo-api-1",
    "entity_uid": "container:stack-demo-api-1",
    "service": "demo-api",
    "state": "running",
    "health": "healthy",
    "restart_count": 0,
    "oom_killed": false,
    "status_text": "Up 2 hours",
    "updated_at": "2026-09-04T12:00:00Z"
  }
]
```

**Response** `202` sem corpo. `exit_code` é omitido quando zero.

## POST `/api/v1/events`

Um `Event`. `id` vazio recebe UUID; `ts` vazio recebe o horário do hub. A resposta é `202`.

```json
{
  "id": "3c1f0a2b-9d44-4e11-8a77-1b2c3d4e5f60",
  "type": "container.oom",
  "ts": "2026-09-04T12:00:00Z",
  "severity": "critical",
  "source": "agent",
  "entity_uid": "docker:docker-host:stack-demo-api-1",
  "labels": { "host": "docker-host", "container": "stack-demo-api-1" },
  "payload": { "action": "oom", "container_id": "abc123" }
}
```

**Response** `202`

```json
{ "id": "3c1f0a2b-9d44-4e11-8a77-1b2c3d4e5f60" }
```

Tipos que o agent emite a partir de eventos Docker: `container.start`, `container.die`, `container.stop`, `container.kill`, `container.oom`, `container.restart`, `container.pause`, `container.unpause`, `container.destroy`, `container.rename`. O payload de `container.die` inclui `cause`: `intentional` (saída 0 ou sinal de parada) ou `unexpected`. O hub também publica no mesmo endpoint interno do bus (`agent.register`, `alert.fired`, `metric.threshold`, `alert.resolved`, `slo.budget_low`).

## POST `/v1/traces`

Única rota OTLP do hub. JSON (`application/json`), corpo até 4 MiB. Handler: `internal/hub/otlp.go`. Parser: `internal/otel/ingest.go`.

Não há `POST /v1/logs` nem `POST /v1/metrics`. Protobuf OTLP não é aceito.

O parser lê `resourceSpans[].resource.attributes` (`service.name`, `container.id` ou `container.name`) e `scopeSpans[].spans`. IDs em hex de 32 caracteres, UUID ou base64. `startTimeUnixNano` e `endTimeUnixNano` são strings decimais. `kind` é o inteiro OTLP (1 internal, 2 server, 3 client, 4 producer, 5 consumer). `status.code` 1 vira `ok`, 2 vira `error`.

```json
{
  "resourceSpans": [
    {
      "resource": {
        "attributes": [
          { "key": "service.name", "value": { "stringValue": "demo-api" } },
          { "key": "container.name", "value": { "stringValue": "stack-demo-api-1" } }
        ]
      },
      "scopeSpans": [
        {
          "spans": [
            {
              "traceId": "5b8efff798038103d269b633813fc60c",
              "spanId": "7a1b2c3d4e5f6071",
              "name": "POST /checkout",
              "kind": 2,
              "startTimeUnixNano": "1759169798309000000",
              "endTimeUnixNano": "1759169798949000000",
              "attributes": [
                { "key": "http.route", "value": { "stringValue": "/checkout" } }
              ],
              "status": { "code": 1 }
            }
          ]
        }
      ]
    }
  ]
}
```

**Response** `200` sem corpo. JSON inválido: `400` `invalid otlp json`. Span sem `traceId` ou `spanId` decodificável é ignorado; se nenhum span restar, a resposta continua `200` e nada é gravado.

## Códigos de erro

| Código | Situação |
|---|---|
| `401` | Token ausente ou inválido |
| `400` | JSON inválido |
| `500` | Erro de persistência |
