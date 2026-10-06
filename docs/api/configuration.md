# Configuration

Copy `.env.example` to `.env`. Empty values use the binary default. Durations use `time.ParseDuration` (`15s`, `1h`, `168h`).

Recreate containers after changing `.env` (`docker compose up -d --force-recreate`).

## Hub

| Variable | Description | Default | Required | Example |
|---|---|---|---|---|
| `ARGUS_HUB_ADDR` | HTTP bind address | `:8080` | no | `:8080` |
| `ARGUS_STORE_DRIVER` | Store driver. `postgres` is built in. A document driver registers with `factory.Register` and is selected by name | `postgres` | no | `postgres` |
| `ARGUS_STORE_DSN` | Driver connection string. Required for `postgres` | empty | yes | `postgres://argus:change-me@postgres:5432/argus?sslmode=disable` |
| `ARGUS_AGENT_TOKEN` | Bearer required on ingest routes. Empty disables the check | empty | no | `change-me` |
| `ARGUS_RETENTION_LOGS` | Max age of `log_entries` | `168h` | no | `168h` |
| `ARGUS_RETENTION_METRICS` | Max age of `metric_points` | `720h` | no | `720h` |
| `ARGUS_RETENTION_EVENTS` | Max age of `events` | `720h` | no | `720h` |
| `ARGUS_PURGE_INTERVAL` | Purge job period | `1h` | no | `1h` |
| `ARGUS_PURGE_TIMEOUT` | Timeout of one purge run | `5s` | no | `5s` |
| `ARGUS_INGEST_CONCURRENCY` | Ingest requests processed at once (all Bearer-authenticated routes except register/heartbeat) | `8` | no | `8` |
| `ARGUS_INGEST_WAIT` | Max wait for an ingest slot; then `503` with `Retry-After: 5` | `2s` | no | `2s` |
| `ARGUS_MAX_BODY_BYTES` | Max body size of an ingest request | `8388608` | no | `8388608` |

Agents on other hosts: `argus-web` also listens on port `8081`, which only proxies ingest POSTs and `/health` to the hub. Publish that port on the reverse proxy (not `80`, and not the hub directly: hub GET routes have no auth) and set `ARGUS_AGENT_TOKEN` on the hub. Each agent needs its own `ARGUS_AGENT_ID` and `ARGUS_HOST_ID`.

The hub opens one driver for every table. Postgres is the relational driver. Another backend, including a document store, implements `store.Store` and registers under its own name; `ARGUS_STORE_DRIVER` selects it. There is no embedded database and no migration from a previous file.

Purge also removes `log_patterns` (log retention) and `topology_edges` (metric retention) by `last_seen`.

An invalid duration falls back to the default. An empty `ARGUS_STORE_DSN` makes the hub exit at startup.

## Agent

| Variable | Description | Default | Required | Example |
|---|---|---|---|---|
| `ARGUS_HUB_URL` | Hub base URL, no trailing slash | `http://127.0.0.1:8080` | no | `http://argus-hub:8080` |
| `ARGUS_AGENT_TOKEN` | Same value as the hub when the hub requires Bearer | empty | no | `change-me` |
| `ARGUS_AGENT_ID` | Agent process identifier | hostname | no | `agent-1` |
| `ARGUS_HOST_ID` | Observed host identifier | hostname | no | `docker-host` |
| `ARGUS_COLLECT_INTERVAL` | Docker metrics collection interval | `15s` | no | `15s` |

The agent also reads variables that `.env.example` does not yet list:

| Variable | Description | Default | Required | Example |
|---|---|---|---|---|
| `ARGUS_LOG_INTERVAL` | Log collection interval | `30s` | no | `30s` |
| `ARGUS_FLEET_INTERVAL` | Fleet snapshot interval | `60s` | no | `60s` |
| `ARGUS_NAME_PREFIX` | Only containers whose name starts with the prefix. Empty collects all | empty | no | `stack-` |
| `DOCKER_HOST` | Docker Unix socket. `unix://` only | `unix:///var/run/docker.sock` | no | `unix:///var/run/docker.sock` |
| `ARGUS_EBPF` | Enable the kernel collector (`1`, `true`, `yes`, `on`) | off | no | `1` |
| `ARGUS_EBPF_INTERVAL` | Kernel collection window | `30s` | no | `30s` |

### Kernel collector

With `ARGUS_EBPF` on, the agent attaches two kernel tracepoints (`sock/inet_sock_set_state` and `tcp/tcp_retransmit_skb`) and reports observed dependencies plus connection counts, instead of relying on applications to log them. See [ingest.md](ingest.md).

The agent container needs `cap_bpf`, `cap_perfmon` and `cap_sys_admin` (Docker drops the cgroup-level permission that the fine-grained capabilities rely on), a read-only mount of `/sys/kernel/tracing`, and the host kernel must expose BTF. When either is missing the agent logs the reason once and keeps running on the log-derived pipeline. With the flag off, the agent needs no extra privileges and behaves exactly as before.

## Host identifier

`ARGUS_AGENT_ID` and `ARGUS_HOST_ID` are labels chosen by whoever operates the agent. They do not come from inventory, a provider, or a machine name of the install.

| Variable | Role |
|---|---|
| `ARGUS_AGENT_ID` | Identifies the process. Goes in `agent_id` on register and heartbeat |
| `ARGUS_HOST_ID` | Identifies the host (or Docker environment) the agent observes. Enters the `host` label and the `entity_uid` of events (`docker:<host_id>:<container>`) |

When the variable is empty, the binary uses the system hostname. If the hostname cannot be read, the fallback is `unknown`.

Use a stable pair across restarts (`agent-1`, `docker-host`) so metrics, logs, and events stay on the same host. The example value in `.env.example` is that generic label, not a specific host.

## Web (dev)

The production panel talks to the hub through the container proxy. These variables apply to `npm run dev`:

| Variable | Description | Default | Required | Example |
|---|---|---|---|---|
| `VITE_API_BASE` | API prefix in the browser. Empty uses the same origin (Vite proxy) | empty | no | `` |
| `VITE_HUB_PROXY` | Development proxy upstream (`/api` and `/health`) | `http://127.0.0.1:8080` | no | `http://127.0.0.1:8080` |

Neither is in `.env.example`.

## Docker Compose

The compose file at the repo root publishes:

- Hub: `8080:8080`
- Panel: `3000:80`
