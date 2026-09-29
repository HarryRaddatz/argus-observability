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

Grafo de dependências inferido a partir de logs (`TopologyGraph` em `internal/model/types.go`).

| Parâmetro | Default | Descrição |
|---|---|---|
| `since` | `24h` | Janela (`time.ParseDuration`). Valor inválido mantém o default |

`kind` observado na inferência: `http`, `amqp`. Listas vazias vêm como `[]`, não `null`.

**Exemplo:** `GET /api/v1/topology?since=24h`

```json
{
  "nodes": [
    { "id": "demo-api", "label": "demo-api" },
    { "id": "checkout", "label": "checkout" }
  ],
  "edges": [
    {
      "source": "checkout",
      "target": "demo-api",
      "kind": "http",
      "count": 12
    }
  ]
}
```

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

Alertas ainda abertos no rule engine (`internal/rules/engine.go`, serializados em `internal/hub/topology.go`). Sem query. Se o engine não estiver ligado, o corpo é `[]`.

Regras embutidas: `cpu-high` (`cpu.usage` ≥ 80 por 5 minutos, `warning`) e `memory-high` (`memory.usage_pct` ≥ 90 por 5 minutos, `critical`).

```json
[
  {
    "rule_id": "cpu-high",
    "entity_uid": "container:stack-demo-api-1",
    "container": "stack-demo-api-1",
    "title": "CPU elevada — stack-demo-api-1",
    "severity": "warning",
    "summary": "cpu.usage em 82.1 (limiar 80)",
    "fired_at": "2026-09-29T18:10:00Z",
    "value": 82.1
  }
]
```

## SLOs

Contrato: `SLODefinition` e `SLOStatus` em `internal/model/types.go`. Handlers em `internal/hub/slos.go`.

| Rota | Corpo |
|---|---|
| GET `/api/v1/slos` | Array de definições. Vazio: `[]` |
| GET `/api/v1/slos/status` | Array de status. Definição que falha na avaliação é omitida |
| GET `/api/v1/slos/{id}/status` | Um status. `404` com texto `not found` se o id não existe |

Seed inclui dois SLOs do serviço `demo-api`, target `99.9`, janela `720` horas: `slo-demo-latency` (`latency_p95`, limiar 500 ms) e `slo-demo-availability` (`availability`).

**Exemplo:** `GET /api/v1/slos/status`

```json
[
  {
    "slo": {
      "id": "slo-demo-availability",
      "name": "Disponibilidade demo-api",
      "service": "demo-api",
      "sli_metric": "availability",
      "target": 99.9,
      "window_hours": 720,
      "latency_threshold_ms": 0,
      "created_at": "2026-09-01T00:00:00Z"
    },
    "compliance": 99.95,
    "error_budget_remaining": 50,
    "good_events": 1999,
    "total_events": 2000,
    "p95_latency_ms": 180,
    "breached": false,
    "evaluated_at": "2026-09-29T18:16:00Z"
  }
]
```

`GET /api/v1/slos/{id}/status` devolve o mesmo objeto, sem o array. `group_id` só aparece quando a definição tem grupo.

## Pacotes relacionados

- `internal/insights/` — classificação, métricas HTTP, insights
- `internal/rules/` — regras CPU/mem/SLO
- `internal/slo/` — avaliador de SLO
- `internal/traces/` — builder a partir de logs + OTLP store
