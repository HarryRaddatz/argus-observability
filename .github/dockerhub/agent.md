# argus-agent

Agent do [Argus](https://github.com/HarryRaddatz/argus-observability): coleta CPU, memória, rede, block I/O e logs dos containers Docker do host e envia ao `argus-hub`.

Imagens: `linux/amd64`, `linux/arm64` · Base `alpine:3.24` · Licença MIT · [Documentação](https://harryraddatz.github.io/argus-observability/)

Também publicada no GHCR: `ghcr.io/harryraddatz/argus-agent`.

## Uso rápido

Stack completa: veja [`argus-hub`](https://hub.docker.com/r/pseudohuery/argus-hub).

Só o agent, apontando para um hub existente:

```bash
docker run -d --name argus-agent \
  -e ARGUS_HUB_URL=http://argus-hub:8080 \
  -e ARGUS_HOST_ID=my-host \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  pseudohuery/argus-agent:latest
```

O agent precisa ler o socket do Docker. Se o socket do host não for legível pelo usuário do container, adicione o grupo dono do socket com `--group-add $(stat -c %g /var/run/docker.sock)`.

## Tags

| Tag | Conteúdo |
|---|---|
| `X.Y.Z`, `vX.Y.Z` | Release exata (recomendado em produção) |
| `X.Y` | Último patch da linha `X.Y` |
| `X` | Última release da major `X` (a partir de `1.0`) |
| `latest` | Última release estável |

Use a mesma versão no hub e no agent.

## Configuração

| Variável | Default | Descrição |
|---|---|---|
| `ARGUS_HUB_URL` | `http://127.0.0.1:8080` | URL do hub |
| `ARGUS_AGENT_TOKEN` | vazio | Token de ingest (igual ao do hub) |
| `ARGUS_AGENT_ID` | hostname do container | Identificador do agent |
| `ARGUS_HOST_ID` | hostname do container | Nome do host nos painéis; defina para não mudar a cada recriação |
| `ARGUS_COLLECT_INTERVAL` | `15s` | Intervalo de coleta de métricas |
| `ARGUS_LOG_INTERVAL` | `30s` | Intervalo de coleta de logs |
| `ARGUS_FLEET_INTERVAL` | `60s` | Intervalo de envio do inventário do host |
| `ARGUS_NAME_PREFIX` | vazio | Coleta só containers cujo nome começa com o prefixo; vazio coleta todos |

Referência completa: [configuração](https://harryraddatz.github.io/argus-observability/api/configuration/).

## Links

- Código e issues: [GitHub](https://github.com/HarryRaddatz/argus-observability)
- Novidades: [changelog](https://harryraddatz.github.io/argus-observability/changelog/)
