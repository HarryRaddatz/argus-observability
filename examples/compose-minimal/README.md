# Minimal example — Argus

Reduced stack to validate ingest and the panel.

## Local build (default)

From the repository root:

```bash
docker compose -f examples/compose-minimal/docker-compose.yml up -d --build
```

## Published images (GHCR)

After a [release](https://github.com/HarryRaddatz/argus-observability/releases) on GitHub:

```bash
docker compose -f examples/compose-minimal/docker-compose.published.yml pull
docker compose -f examples/compose-minimal/docker-compose.published.yml up -d
```

Specific version:

```bash
ARGUS_VERSION=0.1.1 docker compose -f examples/compose-minimal/docker-compose.published.yml up -d
```

| Service | URL |
|---|---|
| Panel | http://localhost:3000 |
| Hub | http://localhost:8080/health |

## Check ingest

```bash
curl -s http://localhost:8080/health
curl -s -H "Authorization: Bearer $ARGUS_AGENT_TOKEN" 'http://localhost:8080/api/v1/workloads?since=30m' | head -c 500
```

## Stop

```bash
docker compose -f examples/compose-minimal/docker-compose.yml down
# or
docker compose -f examples/compose-minimal/docker-compose.published.yml down
```

The build compose uses the root context (`../../`) — clone the full repo before you start it.
