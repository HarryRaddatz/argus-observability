# argus-web

Painel web do [Argus](https://github.com/HarryRaddatz/argus-observability): dashboard, workloads, métricas, logs, traces, SLOs, topologia e eventos. Build estático servido por nginx.

Imagens: `linux/amd64`, `linux/arm64` · Licença MIT · [Documentação](https://harryraddatz.github.io/argus-observability/)

## Uso rápido

Stack completa: veja [`argus-hub`](https://hub.docker.com/r/pseudohuery/argus-hub).

O nginx da imagem faz proxy de `/api/` e `/health` para `http://argus-hub:8080`. O container precisa estar na mesma rede Docker de um hub chamado `argus-hub`:

```bash
docker network create argus
docker run -d --name argus-hub --network argus -v argus_data:/data \
  -e ARGUS_STORE_PATH=/data/argus.db pseudohuery/argus-hub:latest
docker run -d --name argus-web --network argus -p 3000:80 pseudohuery/argus-web:latest
```

Painel em `http://localhost:3000`.

## Tags

| Tag | Conteúdo |
|---|---|
| `X.Y.Z` | Release exata (recomendado em produção) |
| `X.Y` | Último patch da linha `X.Y` |
| `latest` | Última release estável |

Use a mesma versão no hub e no painel.

## Links

- Código e issues: [GitHub](https://github.com/HarryRaddatz/argus-observability)
- Guia do painel: [rotas e navegação](https://harryraddatz.github.io/argus-observability/flows/ui-panel/)
