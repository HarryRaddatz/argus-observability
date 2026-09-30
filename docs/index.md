# Argus

<p class="lead">Observability for Docker hosts and workloads: metrics, logs, events, traces, and a web panel. A central hub plus a light agent per host. Persistence is SQLite by default.</p>

[Deploy](deploy.md){ .md-button .md-button--primary }
[API](api/README.md){ .md-button }

<div class="grid cards" markdown>

-   **Overview**

    ---

    How the agent, hub, and store fit together.

    [Architecture](overview.md)

-   **Deploy**

    ---

    Upgrade with published images or a local build.

    [Runbook](deploy.md)

-   **Panel**

    ---

    Seven destinations. Filters live in the URL. English and Portuguese.

    [Routes](flows/ui-panel.md)

-   **API**

    ---

    Ingest, query, and derived observability.

    [Reference](api/README.md)

</div>

## Quickstart

Docker Engine with Compose v2:

```bash
git clone https://github.com/HarryRaddatz/argus-observability.git
cd argus-observability
cp .env.example .env
docker compose up -d --build
```

| Service | URL |
|---|---|
| Panel | http://localhost:3000 |
| Hub API | http://localhost:8080 |
| Health | http://localhost:8080/health |

Images: [GitHub Releases](https://github.com/HarryRaddatz/argus-observability/releases) · Docker Hub `pseudohuery/argus-{hub,agent,web}` · GHCR `ghcr.io/harryraddatz/argus-{hub,agent,web}`.

The GitHub [README](https://github.com/HarryRaddatz/argus-observability#readme) has clone and contribute links. Portuguese: [Português](/pt/).
