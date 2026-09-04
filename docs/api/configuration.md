# Configuração

Copie `.env.example` para `.env`.

## Hub

| Variável | Default | Descrição |
|---|---|---|
| `ARGUS_HUB_ADDR` | `:8080` | Endereço de bind |
| `ARGUS_STORE_PATH` | `/data/argus.db` | Caminho SQLite |
| `ARGUS_AGENT_TOKEN` | — | Token Bearer para ingest (opcional em dev) |
| `ARGUS_RETENTION_LOGS` | `168h` | Retenção de logs |
| `ARGUS_RETENTION_METRICS` | `720h` | Retenção de métricas |
| `ARGUS_RETENTION_EVENTS` | `720h` | Retenção de eventos |
| `ARGUS_PURGE_INTERVAL` | `1h` | Intervalo do job de purge |
| `ARGUS_PURGE_TIMEOUT` | `5s` | Timeout por execução de purge |

## Agent

| Variável | Default | Descrição |
|---|---|---|
| `ARGUS_HUB_URL` | — | URL base do hub (ex.: `http://argus-hub:8080`) |
| `ARGUS_AGENT_ID` | `argus-agent` | Identificador do agent |
| `ARGUS_HOST_ID` | `docker-host` | Identificador lógico do host |
| `ARGUS_COLLECT_INTERVAL` | `15s` | Intervalo entre coletas Docker |

## Web (dev)

| Variável | Default | Descrição |
|---|---|---|
| `VITE_API_BASE` | `` | Prefixo API (vazio = proxy Vite) |
| `VITE_HUB_PROXY` | `http://127.0.0.1:8080` | Upstream do proxy dev |

## Docker Compose

O compose na raiz publica:

- Hub: `8080:8080`
- Painel: `3000:80`

Recrie containers após alterar `.env` (`docker compose up -d --force-recreate`).
