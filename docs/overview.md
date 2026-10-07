# Overview

Argus splits **collection** (agent) from **aggregation and rules** (hub). Everything that moves between them uses stable envelopes with labels and `entity_uid`, so metrics, logs, and events share the same filter vocabulary.

## Topology

```mermaid
flowchart TB
  subgraph hostA [Host A]
    A1[Agent]
    D1[Docker runtime]
    A1 --> D1
  end
  subgraph hostB [Host B]
    A2[Agent]
    D2[Docker runtime]
    A2 --> D2
  end
  Hub[Hub]
  Store[(Store)]
  A1 --> Hub
  A2 --> Hub
  Hub --> Store
```

## Hub layers

```mermaid
flowchart TB
  Ingest[Ingest API]
  WS[WebSocket]
  Bus[Event bus]
  Rules[Rule engine]
  Store[(Store)]
  Ingest --> Store
  Ingest --> Bus
  Bus --> Rules
  Rules --> Store
  Rules --> Notify[Notifiers]
  WS --> Store
```

## Workload entity

Every signal points at an entity:

```
entity_uid = {runtime}:{host}:{workload_id}
```

Example: `docker:vps-01:api-gateway`

Common labels: `host`, `runtime`, `container`, `namespace`, `pod`, `service`.

## Agent states

```mermaid
stateDiagram-v2
  [*] --> Starting
  Starting --> Connected: register ok
  Connected --> Retrying: 429, 5xx, or network
  Retrying --> Connected: 2xx
  Retrying --> Spooling: five attempts failed
  Spooling --> Connected: oldest batch accepted
  Connected --> Connected: 401, 403, or other 4xx drops that batch
```

Register and heartbeat are not written to the spool. A full spool drops the oldest batch (`ARGUS_BUFFER_MAX_BYTES`).

## Next flows

- [agent-connection.md](flows/agent-connection.md)
- [metrics-ingestion.md](flows/metrics-ingestion.md)
- [log-streaming.md](flows/log-streaming.md)
- [events-and-alerts.md](flows/events-and-alerts.md)
