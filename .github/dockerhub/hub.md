# argus-hub

Central hub of [Argus](https://github.com/HarryRaddatz/argus-observability): REST API, ingest of metrics, logs, events and OTLP traces, rules and SLOs. Postgres persistence.

Images: `linux/amd64`, `linux/arm64` · Base `alpine:3.24` · MIT license · [Documentation](https://harryraddatz.github.io/argus-observability/)

Also published on GHCR: `ghcr.io/harryraddatz/argus-hub`.

## Quick start

The full stack (hub, agent, and panel) comes up with the repository compose:

```bash
curl -fsSLO https://raw.githubusercontent.com/HarryRaddatz/argus-observability/main/examples/compose-minimal/docker-compose.published.yml
ARGUS_REGISTRY=docker.io/pseudohuery docker compose -f docker-compose.published.yml up -d
```

Hub only:

```bash
docker run -d --name argus-hub -p 8080:8080 \
  -e ARGUS_STORE_DRIVER=postgres \
  -e ARGUS_STORE_DSN=postgres://argus:change-me@postgres:5432/argus?sslmode=disable \
  pseudohuery/argus-hub:latest

curl -s http://localhost:8080/health
```

## Tags

| Tag | Content |
|---|---|
| `X.Y.Z`, `vX.Y.Z` | Exact release (recommended in production) |
| `X.Y` | Latest patch of the `X.Y` line |
| `X` | Latest release of major `X` (from `1.0`) |
| `latest` | Latest stable release |

Pre-releases (`X.Y.Z-rc.N`) do not move `latest`.

## Configuration

| Variable | Default | Description |
|---|---|---|
| `ARGUS_HUB_ADDR` | `:8080` | Listen address |
| `ARGUS_STORE_DRIVER` | `postgres` | Store driver. A registered document driver is selected by name |
| `ARGUS_STORE_DSN` | empty | Connection string for the selected driver |
| `ARGUS_AGENT_TOKEN` | empty | Token required on ingest; empty disables authentication |
| `ARGUS_RETENTION_LOGS` | `168h` | Log retention |
| `ARGUS_RETENTION_METRICS` | `720h` | Metric retention |
| `ARGUS_RETENTION_EVENTS` | `720h` | Event retention |
| `ARGUS_PURGE_INTERVAL` | `1h` | Purge job interval |
| `ARGUS_PURGE_TIMEOUT` | `5s` | Timeout per purge run |
| `ARGUS_INGEST_CONCURRENCY` | `8` | Concurrent ingest requests |
| `ARGUS_INGEST_WAIT` | `2s` | Wait for an ingest slot before `503` |
| `ARGUS_MAX_BODY_BYTES` | `8388608` | Max ingest body size |

Full reference: [configuration](https://harryraddatz.github.io/argus-observability/api/configuration/).

Query routes use the same Bearer token as ingest when `ARGUS_AGENT_TOKEN` is set. `/health` stays open. Do not expose port `8080` to the internet; put the panel in front of it.

## Links

- Code and issues: [GitHub](https://github.com/HarryRaddatz/argus-observability)
- What's new: [changelog](https://harryraddatz.github.io/argus-observability/changelog/)
- Upgrade an instance: [deploy](https://harryraddatz.github.io/argus-observability/deploy/)
