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

Padrões normalizados de mensagens de log, do mais frequente para o menos (até 50).

| Parâmetro | Default | Descrição |
|---|---|---|
| `since` | `1h` | Janela (`time.ParseDuration`) |
| `container` | — | Só padrões deste container |
| `group` | — | Só containers do grupo; `404` se o grupo não existir |
| `q` | — | Texto contido no padrão ou na amostra (sem diferenciar maiúsculas) |

**Exemplo:** `GET /api/v1/logs/patterns?since=6h&group=stack-shop&q=timeout`

## GET `/api/v1/topology`

Grafo de dependências inferido a partir de logs.

## GET `/api/v1/traces`

Traces recentes, do mais novo para o mais antigo. Junta spans OTLP e linhas de log com `trace_id`; um trace presente nas duas fontes aparece uma vez, pelos spans OTLP.

| Parâmetro | Default | Descrição |
|---|---|---|
| `since` | `1h` | Janela (`time.ParseDuration`) |
| `service` | — | Traces com pelo menos um span ou linha deste serviço |
| `limit` | `50` | Máximo `200` |

Cada fonte lê no máximo 5000 linhas da janela.

**Exemplo:** `GET /api/v1/traces?since=1h&service=demo-api`

```json
[
  {
    "trace_id": "5b8efff798038103d269b633813fc60c",
    "source": "otlp",
    "service": "checkout",
    "container": "shop-api-1",
    "name": "POST /checkout",
    "start_ts": "2026-09-29T18:16:38.309Z",
    "end_ts": "2026-09-29T18:16:38.949Z",
    "duration_ms": 640,
    "span_count": 2,
    "error": true
  }
]
```

`source` é `otlp` ou `logs`. Em traces de logs, `name` vem de `route`, `path`, `msg` ou `event` quando a linha é JSON, e `span_count` conta as linhas.

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
