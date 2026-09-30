# Argus

<p class="lead">Observabilidade de hosts e workloads Docker: métricas, logs, eventos, traces e painel web. Hub central e um agent leve por host. Persistência em SQLite no default.</p>

[Deploy](deploy.md){ .md-button .md-button--primary }
[API](api/README.md){ .md-button }

<div class="grid cards" markdown>

-   **Visão geral**

    ---

    Como agent, hub e store se encaixam.

    [Arquitetura](overview.md)

-   **Deploy**

    ---

    Atualize com imagens publicadas ou build local.

    [Runbook](deploy.md)

-   **Painel**

    ---

    Sete destinos. Filtros na URL. Inglês e português.

    [Rotas](flows/ui-panel.md)

-   **API**

    ---

    Ingest, consulta e observabilidade derivada.

    [Referência](api/README.md)

</div>

## Quickstart

Docker Engine com Compose v2:

```bash
git clone https://github.com/HarryRaddatz/argus-observability.git
cd argus-observability
cp .env.example .env
docker compose up -d --build
```

| Serviço | URL |
|---|---|
| Painel | http://localhost:3000 |
| Hub API | http://localhost:8080 |
| Health | http://localhost:8080/health |

Imagens: [GitHub Releases](https://github.com/HarryRaddatz/argus-observability/releases) · Docker Hub `pseudohuery/argus-{hub,agent,web}` · GHCR `ghcr.io/harryraddatz/argus-{hub,agent,web}`.

O [README](https://github.com/HarryRaddatz/argus-observability#readme) no GitHub tem clone e contribuição. English: [English](/).
