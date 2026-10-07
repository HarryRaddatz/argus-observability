# argus-web

[Argus](https://github.com/HarryRaddatz/argus-observability) web panel: overview, containers, metrics, logs, traces, SLOs, topology, and events. Static build served by nginx.

Images: `linux/amd64`, `linux/arm64` · Base `nginx:1.30-alpine` · MIT license · [Documentation](https://harryraddatz.github.io/argus-observability/)

Also published on GHCR: `ghcr.io/harryraddatz/argus-web`.

## Quick start

Full stack: see [`argus-hub`](https://hub.docker.com/r/pseudohuery/argus-hub).

The image nginx proxies `/api/` and `/health` to `http://argus-hub:8080`. The container must share a Docker network with a hub named `argus-hub`:

```bash
docker network create argus
docker run -d --name argus-hub --network argus \
  -e ARGUS_STORE_DSN=postgres://argus:change-me@postgres:5432/argus?sslmode=disable \
  pseudohuery/argus-hub:latest
docker run -d --name argus-web --network argus -p 3000:80 pseudohuery/argus-web:latest
```

Panel at `http://localhost:3000`.

## Tags

| Tag | Content |
|---|---|
| `X.Y.Z`, `vX.Y.Z` | Exact release (recommended in production) |
| `X.Y` | Latest patch of the `X.Y` line |
| `X` | Latest release of major `X` (from `1.0`) |
| `latest` | Latest stable release |

Use the same version on the hub and the panel.

## Links

- Code and issues: [GitHub](https://github.com/HarryRaddatz/argus-observability)
- Panel guide: [routes and navigation](https://harryraddatz.github.io/argus-observability/flows/ui-panel/)
