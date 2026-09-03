# Fluxo: eventos e alertas

Eventos tipados circulam no bus interno. Regras avaliam condições e disparam notificações com deduplicação.

## Envelope de evento

```json
{
  "id": "evt_01J...",
  "type": "metric.threshold",
  "ts": "2026-09-03T13:00:00Z",
  "severity": "warning",
  "source": "rule-engine",
  "entity_uid": "docker:host-01:api",
  "labels": { "host": "host-01", "rule": "cpu-high" },
  "payload": { "metric": "cpu.usage", "value": 92.1, "threshold": 90 }
}
```

## Sequência — publicação no bus

```mermaid
sequenceDiagram
  participant Source as Origem
  participant Bus
  participant Store
  participant Rules
  participant Notify

  Source->>Bus: Publish(event)
  Bus->>Store: persist async
  Bus->>Rules: dispatch
  alt regra match
    Rules->>Bus: alert.fired
    Rules->>Notify: send
  end
```

## Sequência — regra de limiar

```mermaid
sequenceDiagram
  participant Metrics as Ingest metricas
  participant Rules
  participant Store
  participant Notify

  Metrics->>Rules: tick evaluate
  Rules->>Store: query avg last 5m
  Store-->>Rules: value
  alt value > threshold for 2m
    Rules->>Store: insert event metric.threshold
    Rules->>Notify: discord / webhook / ntfy
  else recovered
    Rules->>Store: insert event alert.resolved
  end
```

## Deduplicação

```mermaid
flowchart LR
  Event[Evento candidato] --> Key{dedupe_key existe?}
  Key -->|sim, janela ativa| Drop[Ignora]
  Key -->|nao| Fire[Dispara alerta]
  Fire --> Cache[Registra janela]
```

## Catálogo de tipos (v1)

| type | severity típica |
|---|---|
| `agent.disconnect` | warning |
| `agent.backpressure` | warning |
| `metric.threshold` | warning / critical |
| `resource.pressure` | warning |
| `container.oom` | critical |
| `container.restart` | info |
| `log.pattern` | warning |
| `alert.fired` | info |
| `alert.resolved` | info |
| `notify.failed` | warning |

## Timeline na UI

```mermaid
sequenceDiagram
  participant UI
  participant Hub
  participant Store

  UI->>Hub: GET /api/v1/events?entity_uid=...&since=1h
  Hub->>Store: list by ts desc
  Store-->>Hub: events
  Hub-->>UI: timeline JSON
```

## Notificações

Fan-out assíncrono com retry. Falha gera `notify.failed` no bus.

Canais v1: webhook genérico (JSON template).
