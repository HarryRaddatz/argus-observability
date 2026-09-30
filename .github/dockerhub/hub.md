# argus-hub

Hub central do [Argus](https://github.com/HarryRaddatz/argus-observability): API REST, ingest de métricas, logs, eventos e traces OTLP, regras e SLOs. Persistência em SQLite.

Imagens: `linux/amd64`, `linux/arm64` · Base `alpine:3.24` · Licença MIT · [Documentação](https://harryraddatz.github.io/argus-observability/)

Também publicada no GHCR: `ghcr.io/harryraddatz/argus-hub`.

## Uso rápido

A stack completa (hub, agent e painel) sobe com o compose do repositório:

```bash
curl -fsSLO https://raw.githubusercontent.com/HarryRaddatz/argus-observability/main/examples/compose-minimal/docker-compose.published.yml
ARGUS_REGISTRY=docker.io/pseudohuery docker compose -f docker-compose.published.yml up -d
```

Só o hub:

```bash
docker run -d --name argus-hub -p 8080:8080 \
  -e ARGUS_STORE_PATH=/data/argus.db \
  -v argus_data:/data \
  pseudohuery/argus-hub:latest

curl -s http://localhost:8080/health
```

## Tags

| Tag | Conteúdo |
|---|---|
| `X.Y.Z`, `vX.Y.Z` | Release exata (recomendado em produção) |
| `X.Y` | Último patch da linha `X.Y` |
| `X` | Última release da major `X` (a partir de `1.0`) |
| `latest` | Última release estável |

Pre-releases (`X.Y.Z-rc.N`) não movem `latest`.

## Configuração

| Variável | Default | Descrição |
|---|---|---|
| `ARGUS_HUB_ADDR` | `:8080` | Endereço de escuta |
| `ARGUS_STORE_PATH` | `./data/argus.db` (`/app/data`) | Caminho do arquivo SQLite; use `/data/argus.db` com um volume em `/data` |
| `ARGUS_AGENT_TOKEN` | vazio | Token exigido no ingest; vazio desativa a autenticação |
| `ARGUS_RETENTION_LOGS` | `168h` | Retenção de logs |
| `ARGUS_RETENTION_METRICS` | `720h` | Retenção de métricas |
| `ARGUS_RETENTION_EVENTS` | `720h` | Retenção de eventos |
| `ARGUS_PURGE_INTERVAL` | `1h` | Intervalo do job de purge |
| `ARGUS_PURGE_TIMEOUT` | `5s` | Timeout por execução de purge |
| `ARGUS_INGEST_CONCURRENCY` | `8` | Requests de ingest simultâneos |
| `ARGUS_INGEST_WAIT` | `2s` | Espera por vaga de ingest antes de `503` |
| `ARGUS_MAX_BODY_BYTES` | `8388608` | Tamanho máximo do corpo de ingest |

Referência completa: [configuração](https://harryraddatz.github.io/argus-observability/api/configuration/).

As rotas de consulta não têm autenticação. Não exponha a porta `8080` à internet sem um proxy com autenticação na frente.

## Links

- Código e issues: [GitHub](https://github.com/HarryRaddatz/argus-observability)
- Novidades: [changelog](https://harryraddatz.github.io/argus-observability/changelog/)
- Atualizar uma instância: [deploy](https://harryraddatz.github.io/argus-observability/deploy/)
