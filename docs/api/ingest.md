# Ingest API

POST routes used by the agent. Auth via `Authorization: Bearer` when configured.

## POST `/api/v1/agents/register`

Registers the agent on the hub.

**Request** (`agent_id` and `host_id` required; empty `runtime` becomes `docker`)

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

`interval_seconds` is the hub heartbeat interval (default 30). It is not `ARGUS_COLLECT_INTERVAL`.

## POST `/api/v1/agents/heartbeat`

The handler reads only `agent_id`.

```json
{ "agent_id": "agent-1" }
```

**Response** `200`

```json
{ "status": "ok" }
```

## POST `/api/v1/metrics/batch`

Body: array of `MetricPoint`. An empty array responds `202` without persisting.

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

**Response** `202` with no body.

Infrastructure metrics collected by the agent: `cpu.usage`, `memory.usage`, `memory.limit`, `network.rx`, `network.tx`, `block.read`, `block.write`.

## POST `/api/v1/logs/batch`

Body: array of `LogEntry`. An empty array responds `202` without persisting.

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

**Response** `202` with no body. The hub fills `fields` on ingest; the agent does not need to send them.

## POST `/api/v1/fleet/batch`

Array of `ContainerFleetStatus`. An empty array responds `202` without writing.

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

**Response** `202` with no body. `exit_code` is omitted when zero.

## POST `/api/v1/events`

One `Event`. Empty `id` gets a UUID; empty `ts` gets the hub time. The response is `202`.

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

Types the agent emits from Docker events: `container.start`, `container.die`, `container.oom`, `container.restart`, `container.pause`, `container.unpause`, `container.destroy`, `container.rename`. The hub also publishes on the same internal bus endpoint (`agent.register`, `alert.fired`, `metric.threshold`, `alert.resolved`, `slo.budget_low`).

## POST `/v1/traces`

The hub's only OTLP route. JSON (`application/json`), body up to 4 MiB. Handler: `internal/hub/otlp.go`. Parser: `internal/otel/ingest.go`.

There is no `POST /v1/logs` or `POST /v1/metrics`. OTLP protobuf is not accepted.

The parser reads `resourceSpans[].resource.attributes` (`service.name`, `container.id` or `container.name`) and `scopeSpans[].spans`. IDs as 32-character hex, UUID, or base64. `startTimeUnixNano` and `endTimeUnixNano` are decimal strings. `kind` is the OTLP integer (1 internal, 2 server, 3 client, 4 producer, 5 consumer). `status.code` 1 becomes `ok`, 2 becomes `error`.

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

**Response** `200` with no body. Invalid JSON: `400` `invalid otlp json`. A span without a decodable `traceId` or `spanId` is ignored; if no span remains, the response is still `200` and nothing is written.

## Error codes

| Code | Situation |
|---|---|
| `401` | Missing or invalid token |
| `400` | Invalid JSON |
| `500` | Persistence error |
