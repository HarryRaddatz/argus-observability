# Fluxo: stream de logs

Logs seguem do runtime ao hub e daí para clientes WebSocket. Sob pressão, o agent aplica backpressure antes de saturar o host.

## Sequência — tail contínuo

```mermaid
sequenceDiagram
  participant Runtime
  participant Agent
  participant Hub
  participant Store
  participant UI

  Agent->>Runtime: follow logs (stdout/stderr)
  Runtime-->>Agent: linha + metadata
  Agent->>Agent: parse level, multiline
  Agent->>Hub: POST /api/v1/logs/batch
  Hub->>Store: insert + index
  Hub->>UI: WS broadcast log.entry
```

## Backpressure

```mermaid
flowchart TB
  Prod[Produção de logs] --> Ring[Ring buffer]
  Ring --> Fwd[Forwarder]
  Ring -->|buffer > limite| Sample[Sample 1:N]
  Sample --> Fwd
  Ring -->|crítico| Event[agent.backpressure]
  Event --> Hub
```

| Estágio | Ação |
|---|---|
| buffer < 70% | envio normal |
| 70–90% | sample 1:5 |
| > 90% | sample 1:10 + evento |

## Formato de entrada

```json
{
  "ts": "2026-09-03T13:00:01Z",
  "message": "level=error msg=connection refused",
  "level": "error",
  "entity_uid": "docker:host-01:api",
  "labels": { "host": "host-01", "container": "api" },
  "fields": { "trace_id": "abc" }
}
```

## Sequência — UI tail ao vivo

```mermaid
sequenceDiagram
  participant UI
  participant Hub
  participant Store

  UI->>Hub: WS connect /api/v1/ws/logs?entity_uid=...
  Hub->>Store: subscribe recent + live
  loop live
    Hub-->>UI: log.entry
  end
  UI->>Hub: filter level>=error
  Hub-->>UI: filtered stream
```

## Busca histórica

```mermaid
sequenceDiagram
  participant Client
  participant Hub
  participant Store

  Client->>Hub: GET /api/v1/logs/search?q=refused&since=15m
  Hub->>Store: full-text + labels
  Store-->>Hub: hits
  Hub-->>Client: paginated results
```
