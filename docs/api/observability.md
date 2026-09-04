# Observability API

Endpoints de análise derivada (logs, rules, SLO).

## GET `/api/v1/insights`

Insights automáticos (CPU, memória, HTTP, fleet, patterns, topologia).

**Exemplo:** `GET /api/v1/insights?since=1h`

```json
{
  "since": "2026-09-04T11:00:00Z",
  "insights": [
    {
      "id": "cpu-high-stack-demo-api-1",
      "theme": "cpu",
      "severity": "warning",
      "title": "CPU elevada",
      "summary": "Uso sustentado acima do limiar",
      "container": "stack-demo-api-1",
      "entity_uid": "container:stack-demo-api-1",
      "evidence": { "cpu_usage": 82.1 },
      "recommendations": ["Verificar métricas e logs recentes"]
    }
  ]
}
```

## GET `/api/v1/logs/patterns`

Padrões normalizados de mensagens de log.

## GET `/api/v1/topology`

Grafo de dependências inferido a partir de logs.

## GET `/api/v1/traces/{trace_id}`

Waterfall de spans (OTLP ou reconstruído de logs).

**Exemplo:** `GET /api/v1/traces/50e30959-f59e-487f-bbe0-89ae7d8e74e5?since=24h`

## GET `/api/v1/alerts/active`

Alertas ativos do rule engine (`internal/rules/engine.go`).

## SLOs

| Rota | Descrição |
|---|---|
| GET `/api/v1/slos` | Definições |
| GET `/api/v1/slos/status` | Compliance + error budget |
| GET `/api/v1/slos/{id}/status` | Status de um SLO |

Seed default inclui SLO `demo-api` com target 99.9%.

## Pacotes relacionados

- `internal/insights/` — classificação, métricas HTTP, insights
- `internal/rules/` — regras CPU/mem/SLO
- `internal/slo/` — avaliador de SLO
- `internal/traces/` — builder a partir de logs + OTLP store
