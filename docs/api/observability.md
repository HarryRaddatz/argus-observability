# Observability API

Derived analysis endpoints (logs, rules, SLO).

## GET `/api/v1/insights`

Automatic insights (CPU, memory, HTTP, fleet, patterns, topology).

**Example:** `GET /api/v1/insights?since=1h`

Hub JSON titles stay in the seed language. The panel maps `rule_id` and metric names to the UI locale.

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

Normalized log-message patterns, most frequent first (up to 50).

| Parameter | Default | Description |
|---|---|---|
| `since` | `1h` | Window (`time.ParseDuration`) |
| `container` | — | Only patterns from this container |
| `group` | — | Only containers in the group; `404` if the group does not exist |
| `q` | — | Text contained in the pattern or sample (case-insensitive) |

**Example:** `GET /api/v1/logs/patterns?since=6h&group=stack-shop&q=timeout`

## GET `/api/v1/topology`

Dependency graph (`TopologyGraph` in `internal/model/types.go`). It merges edges observed by the kernel collector with edges inferred from logs. When both describe the same pair, the observed edge wins.

| Parameter | Default | Description |
|---|---|---|
| `since` | `24h` | Window (`time.ParseDuration`). Invalid value keeps the default |

`origin` is `kernel` for observed edges and `log` for inferred ones. Observed edges also carry `port`. `kind` is `http`, `postgres`, `redis`, `amqp`, `mongodb`, `kafka` or `tcp` for observed edges, and `http` or `amqp` for inferred ones. Empty lists come as `[]`, not `null`.

**Example:** `GET /api/v1/topology?since=24h`

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
      "count": 12,
      "origin": "kernel",
      "port": 8080
    }
  ]
}
```

## GET `/api/v1/traces`

Recent traces, newest first. Joins OTLP spans and log lines with `trace_id`; a trace present in both sources appears once, from the OTLP spans.

| Parameter | Default | Description |
|---|---|---|
| `since` | `1h` | Window (`time.ParseDuration`) |
| `service` | — | Traces with at least one span or line from this service |
| `limit` | `50` | Max `200` |

Each source reads at most 5000 rows from the window.

**Example:** `GET /api/v1/traces?since=1h&service=demo-api`

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

`source` is `otlp` or `logs`. In log traces, `name` comes from `route`, `path`, `msg`, or `event` when the line is JSON, and `span_count` counts the lines.

## GET `/api/v1/traces/{trace_id}`

Span waterfall (OTLP or rebuilt from logs).

**Example:** `GET /api/v1/traces/50e30959-f59e-487f-bbe0-89ae7d8e74e5?since=24h`

## GET `/api/v1/alerts/active`

Alerts still open in the rule engine (`internal/rules/engine.go`, serialized in `internal/hub/topology.go`). No query. If the engine is not running, the body is `[]`.

Built-in rules: `cpu-high` (`cpu.usage` ≥ 80 for 5 minutes, `warning`) and `memory-high` (`memory.usage_pct` ≥ 90 for 5 minutes, `critical`).

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

Contract: `SLODefinition` and `SLOStatus` in `internal/model/types.go`. Handlers in `internal/hub/slos.go`.

| Route | Body |
|---|---|
| GET `/api/v1/slos` | Array of definitions. Empty: `[]` |
| GET `/api/v1/slos/status` | Array of statuses. A definition that fails evaluation is omitted |
| GET `/api/v1/slos/{id}/status` | One status. `404` with text `not found` if the id does not exist |

The seed includes two SLOs for service `demo-api`, target `99.9`, window `720` hours: `slo-demo-latency` (`latency_p95`, threshold 500 ms) and `slo-demo-availability` (`availability`).

**Example:** `GET /api/v1/slos/status`

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

`GET /api/v1/slos/{id}/status` returns the same object, without the array. `group_id` only appears when the definition has a group.

## Related packages

- `internal/insights/` — classification, HTTP metrics, insights
- `internal/rules/` — CPU/mem/SLO rules
- `internal/slo/` — SLO evaluator
- `internal/traces/` — builder from logs + OTLP store
