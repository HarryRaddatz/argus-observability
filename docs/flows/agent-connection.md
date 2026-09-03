# Fluxo: conexão agent ↔ hub

O agent inicia a sessão com o hub. Em ambiente local pode usar HTTP direto; em produção, WebSocket TLS com token.

## Sequência — registro e heartbeat

```mermaid
sequenceDiagram
  participant Agent
  participant Hub
  participant Store

  Agent->>Hub: POST /api/v1/agents/register
  Note over Agent,Hub: token, host_id, runtime, labels
  Hub->>Store: upsert agent session
  Hub-->>Agent: 200 session_id, interval

  loop a cada interval
    Agent->>Hub: POST /api/v1/agents/heartbeat
    Hub->>Store: touch last_seen
    Hub-->>Agent: 200 config opcional
  end
```

## Sequência — desconexão detectada

```mermaid
sequenceDiagram
  participant Hub
  participant Store
  participant Bus
  participant Rules

  Hub->>Store: agent last_seen > threshold
  Hub->>Bus: publish agent.disconnect
  Bus->>Rules: evaluate
  Rules->>Store: persist event
  Rules->>Hub: notify channels
```

## Modos de transporte

| Modo | Uso | Endpoint |
|---|---|---|
| HTTP | dev, LAN confiável | `ARGUS_HUB_URL` |
| WebSocket | stream contínuo de logs | `/api/v1/ws` |

## Autenticação

```
Authorization: Bearer <ARGUS_AGENT_TOKEN>
```

Token compartilhado por agent ou por host — configurável no hub.

## Falhas comuns

| Sintoma | Causa provável | Comportamento do agent |
|---|---|---|
| 401 | token inválido | para e loga erro |
| 503 | hub indisponível | backoff exponencial |
| timeout no ingest | hub saturado | backpressure local |
