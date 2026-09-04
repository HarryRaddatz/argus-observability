# Exemplo mínimo — Argus + nginx

Stack reduzida para validar ingest e painel sem dependências externas.

## Uso

Na raiz do repositório:

```bash
docker compose -f examples/compose-minimal/docker-compose.yml up -d --build
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
```

O compose referencia o build da raiz (`../../`) — clone o repo completo antes de subir.
