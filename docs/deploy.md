# Deploy — upgrade the Argus stack

Runbook to promote an Argus version on a Docker host. Applies when you use images published on GHCR or build locally from a clone.

Available versions: [GitHub Releases](https://github.com/HarryRaddatz/argus-observability/releases) · images `ghcr.io/harryraddatz/argus-{hub,agent,web}` and `docker.io/pseudohuery/argus-{hub,agent,web}` (same content, `linux/amd64` and `linux/arm64`).

## Before you upgrade

1. Read the version section in [CHANGELOG.md](../CHANGELOG.md). On a **major** change, look for *Breaking* and adjust `.env` before you start.
2. Confirm `.env` has any new variables listed in [.env.example](../.env.example).
3. The SQLite database lives in the `argus_data` volume. `docker compose down` keeps the volume; **do not** use `down -v`.

Optional database backup (stop the hub first to avoid a torn file):

```bash
docker compose stop argus-hub
docker run --rm -v argus_argus_data:/data -v "$PWD":/backup alpine \
  cp /data/argus.db /backup/argus-$(date +%Y%m%d).db
docker compose start argus-hub
```

The volume name follows `<project>_argus_data`; confirm with `docker volume ls`.

## Option A — published images (GHCR)

Pin the version with `ARGUS_VERSION` (no `v` prefix). Without it, compose uses `latest`. To pull from Docker Hub instead of GHCR, set `ARGUS_REGISTRY=docker.io/pseudohuery`.

```bash
export ARGUS_VERSION=0.1.1
# export ARGUS_REGISTRY=docker.io/pseudohuery
docker compose -f examples/compose-minimal/docker-compose.published.yml pull
docker compose -f examples/compose-minimal/docker-compose.published.yml up -d --force-recreate
```

For a custom compose file, set `image:` to `ghcr.io/harryraddatz/argus-<component>:${ARGUS_VERSION}` and run the same commands.

## Option B — local build

From the clone, on the tag you want:

```bash
git fetch --tags
git checkout v0.1.1
docker compose up -d --build --force-recreate
```

## Why `--force-recreate`

`docker compose restart` restarts the container with the old configuration: it does not reread `env_file` or swap the image. After changing `.env`, a tag, or an image, recreate containers with `up -d --force-recreate`.

## Post-deploy checklist

Adjust host and ports if you are not using the defaults (`8080` hub, `3000` panel).

1. Hub is up:

   ```bash
   curl -fsS http://localhost:8080/health
   # {"status":"ok"}
   ```

2. The panel returns 200 and proxies the API:

   ```bash
   curl -fsS -o /dev/null -w '%{http_code}\n' http://localhost:3000/
   curl -fsS http://localhost:3000/health
   ```

3. Agent is reporting (wait one `ARGUS_COLLECT_INTERVAL` cycle, default 15s):

   ```bash
   curl -fsS 'http://localhost:8080/api/v1/workloads?since=5m' | head -c 500
   curl -fsS http://localhost:8080/api/v1/fleet/status | head -c 500
   ```

   The workload list must not be empty if containers are running on the host.

4. Logs have no ingest authentication errors (token mismatch between hub and agent):

   ```bash
   docker compose logs --since 5m argus-agent | grep -i -E 'error|401' || echo ok
   ```

5. In the panel, **Overview** and **Containers** show recent data.

## Rollback

Go back to the previous version and recreate:

```bash
export ARGUS_VERSION=0.1.0
docker compose -f examples/compose-minimal/docker-compose.published.yml up -d --force-recreate
```

For a local build: `git checkout v0.1.0 && docker compose up -d --build --force-recreate`.

If the new version changed the SQLite schema, restore the backup taken before the deploy.
