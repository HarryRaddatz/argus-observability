# argus-agent

[Argus](https://github.com/HarryRaddatz/argus-observability) agent: collects CPU, memory, network, block I/O, and logs from the host's Docker containers and sends them to `argus-hub`.

Images: `linux/amd64`, `linux/arm64` · Base `alpine:3.24` · MIT license · [Documentation](https://harryraddatz.github.io/argus-observability/)

Also published on GHCR: `ghcr.io/harryraddatz/argus-agent`.

## Quick start

Full stack: see [`argus-hub`](https://hub.docker.com/r/pseudohuery/argus-hub).

Agent only, pointing at an existing hub:

```bash
docker run -d --name argus-agent \
  -e ARGUS_HUB_URL=http://argus-hub:8080 \
  -e ARGUS_HOST_ID=my-host \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  pseudohuery/argus-agent:latest
```

The agent needs to read the Docker socket. If the host socket is not readable by the container user, add the socket owner group with `--group-add $(stat -c %g /var/run/docker.sock)`.

## Tags

| Tag | Content |
|---|---|
| `X.Y.Z`, `vX.Y.Z` | Exact release (recommended in production) |
| `X.Y` | Latest patch of the `X.Y` line |
| `X` | Latest release of major `X` (from `1.0`) |
| `latest` | Latest stable release |

Use the same version on the hub and the agent.

## Configuration

| Variable | Default | Description |
|---|---|---|
| `ARGUS_HUB_URL` | `http://127.0.0.1:8080` | Hub URL |
| `ARGUS_AGENT_TOKEN` | empty | Ingest token (same as the hub) |
| `ARGUS_AGENT_ID` | container hostname | Agent identifier |
| `ARGUS_HOST_ID` | container hostname | Host name in the panels; set it so it does not change on every recreate |
| `ARGUS_COLLECT_INTERVAL` | `15s` | Metrics collection interval |
| `ARGUS_LOG_INTERVAL` | `30s` | Log collection interval |
| `ARGUS_FLEET_INTERVAL` | `60s` | Host inventory send interval |
| `ARGUS_NAME_PREFIX` | empty | Only collect containers whose name starts with the prefix; empty collects all |

Full reference: [configuration](https://harryraddatz.github.io/argus-observability/api/configuration/).

## Links

- Code and issues: [GitHub](https://github.com/HarryRaddatz/argus-observability)
- What's new: [changelog](https://harryraddatz.github.io/argus-observability/changelog/)
