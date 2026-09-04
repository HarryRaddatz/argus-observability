# Exemplo mínimo — Argus

Stack reduzida para validar ingest e painel.

## Build local (default)

Na raiz do repositório:

```bash
docker compose -f examples/compose-minimal/docker-compose.yml up -d --build
```

## Imagens publicadas (GHCR)

Após um [release](https://github.com/HarryRaddatz/argus-observability/releases) no GitHub:

```bash
docker compose -f examples/compose-minimal/docker-compose.published.yml pull
docker compose -f examples/compose-minimal/docker-compose.published.yml up -d
```

Versão específica:

```bash
ARGUS_VERSION=0.1.1 docker compose -f examples/compose-minimal/docker-compose.published.yml up -d
```

| Serviço | URL |
|---|---|
| Painel | http://localhost:3000 |
| Hub | http://localhost:8080/health |

## Verificar ingest

```bash
curl -s http://localhost:8080/health
curl -s 'http://localhost:8080/api/v1/workloads?since=30m' | head -c 500
```

## Parar

```bash
docker compose -f examples/compose-minimal/docker-compose.yml down
# ou
docker compose -f examples/compose-minimal/docker-compose.published.yml down
```

O compose de build referencia o contexto da raiz (`../../`) — clone o repo completo antes de subir.
