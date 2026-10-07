# Visão geral

O Argus separa **coleta** (agent) de **agregação e regras** (hub). Tudo que transita entre eles usa envelopes estáveis com labels e `entity_uid`, para métricas, logs e eventos compartilharem o mesmo vocabulário de filtro.

## Topologia

```mermaid
flowchart TB
  subgraph hostA [Host A]
    A1[Agent]
    D1[Runtime Docker]
    A1 --> D1
  end
  subgraph hostB [Host B]
    A2[Agent]
    D2[Runtime Docker]
    A2 --> D2
  end
  Hub[Hub]
  Store[(Store)]
  A1 --> Hub
  A2 --> Hub
  Hub --> Store
```

## Camadas do hub

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
  Rules --> Notify[Notificadores]
  WS --> Store
```

## Entidade de workload

Cada sinal referencia uma entidade:

```
entity_uid = {runtime}:{host}:{workload_id}
```

Exemplo: `docker:vps-01:api-gateway`

Labels comuns: `host`, `runtime`, `container`, `namespace`, `pod`, `service`.

## Estados do agent

```mermaid
stateDiagram-v2
  [*] --> Starting
  Starting --> Connected: registro ok
  Connected --> Retrying: 429, 5xx ou rede
  Retrying --> Connected: 2xx
  Retrying --> Spooling: cinco tentativas falharam
  Spooling --> Connected: lote mais antigo aceito
  Connected --> Connected: 401, 403 ou outro 4xx descarta esse lote
```

Registro e heartbeat não vão para o spool. Spool cheio descarta o lote mais antigo (`ARGUS_BUFFER_MAX_BYTES`).

## Próximos fluxos

- [agent-connection.md](flows/agent-connection.md)
- [metrics-ingestion.md](flows/metrics-ingestion.md)
- [log-streaming.md](flows/log-streaming.md)
- [events-and-alerts.md](flows/events-and-alerts.md)
