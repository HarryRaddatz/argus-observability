# Configuração

Copie `.env.example` para `.env`. Valores vazios usam o default do binário. Duração no formato de `time.ParseDuration` (`15s`, `1h`, `168h`).

Recrie os containers depois de alterar `.env` (`docker compose up -d --force-recreate`).

## Hub

| Variável | Descrição | Default | Obrigatória | Exemplo |
|---|---|---|---|---|
| `ARGUS_HUB_ADDR` | Endereço de bind HTTP | `:8080` | não | `:8080` |
| `ARGUS_STORE_PATH` | Arquivo SQLite | `./data/argus.db` | não | `/data/argus.db` |
| `ARGUS_AGENT_TOKEN` | Bearer exigido nas rotas de ingest. Vazio desliga a checagem | vazio | não | `change-me` |
| `ARGUS_RETENTION_LOGS` | Idade máxima de `log_entries` | `168h` | não | `168h` |
| `ARGUS_RETENTION_METRICS` | Idade máxima de `metric_points` | `720h` | não | `720h` |
| `ARGUS_RETENTION_EVENTS` | Idade máxima de `events` | `720h` | não | `720h` |
| `ARGUS_PURGE_INTERVAL` | Período do job de purge | `1h` | não | `1h` |
| `ARGUS_PURGE_TIMEOUT` | Timeout de uma execução de purge | `5s` | não | `5s` |

Duração inválida cai no default. O compose e o `.env.example` apontam o SQLite para `/data/argus.db`.

## Agent

| Variável | Descrição | Default | Obrigatória | Exemplo |
|---|---|---|---|---|
| `ARGUS_HUB_URL` | URL base do hub, sem barra final | `http://127.0.0.1:8080` | não | `http://argus-hub:8080` |
| `ARGUS_AGENT_TOKEN` | Mesmo valor do hub, quando o hub exige Bearer | vazio | não | `change-me` |
| `ARGUS_AGENT_ID` | Identificador do processo do agent | hostname | não | `agent-1` |
| `ARGUS_HOST_ID` | Identificador do host observado | hostname | não | `docker-host` |
| `ARGUS_COLLECT_INTERVAL` | Intervalo de coleta de métricas Docker | `15s` | não | `15s` |

O agent também lê variáveis que o `.env.example` ainda não lista:

| Variável | Descrição | Default | Obrigatória | Exemplo |
|---|---|---|---|---|
| `ARGUS_LOG_INTERVAL` | Intervalo de coleta de logs | `30s` | não | `30s` |
| `ARGUS_FLEET_INTERVAL` | Intervalo do snapshot de fleet | `60s` | não | `60s` |
| `ARGUS_NAME_PREFIX` | Só containers cujo nome começa com o prefixo. Vazio coleta todos | vazio | não | `stack-` |
| `DOCKER_HOST` | Socket Unix do Docker. Só `unix://` | `unix:///var/run/docker.sock` | não | `unix:///var/run/docker.sock` |

## Identificador de host

`ARGUS_AGENT_ID` e `ARGUS_HOST_ID` são rótulos escolhidos por quem opera o agent. Não vêm de inventário, provedor ou nome de máquina da instalação.

| Variável | Papel |
|---|---|
| `ARGUS_AGENT_ID` | Identifica o processo. Vai em `agent_id` no registro e no heartbeat |
| `ARGUS_HOST_ID` | Identifica o host (ou o ambiente Docker) que o agent observa. Entra no label `host` e no `entity_uid` dos eventos (`docker:<host_id>:<container>`) |

Com a variável vazia, o binário usa o hostname do sistema. Se o hostname não puder ser lido, o fallback é `unknown`.

Use um par estável entre reinícios (`agent-1`, `docker-host`) para métricas, logs e eventos continuarem no mesmo host. O valor de exemplo do `.env.example` é esse rótulo genérico, não um host específico.

## Web (dev)

O painel em produção fala com o hub pelo proxy do container. Estas variáveis valem para `npm run dev`:

| Variável | Descrição | Default | Obrigatória | Exemplo |
|---|---|---|---|---|
| `VITE_API_BASE` | Prefixo da API no browser. Vazio usa o mesmo origin (proxy do Vite) | vazio | não | `` |
| `VITE_HUB_PROXY` | Upstream do proxy de desenvolvimento (`/api` e `/health`) | `http://127.0.0.1:8080` | não | `http://127.0.0.1:8080` |

Nenhuma das duas está no `.env.example`.

## Docker Compose

O compose na raiz publica:

- Hub: `8080:8080`
- Painel: `3000:80`
