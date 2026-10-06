# Deploy — atualizar a stack Argus

Runbook para promover uma versão do Argus em um docker host. Vale para quem usa as imagens publicadas no GHCR ou faz build local a partir do clone.

Versões disponíveis: [GitHub Releases](https://github.com/HarryRaddatz/argus-observability/releases) · imagens `ghcr.io/harryraddatz/argus-{hub,agent,web}` e `docker.io/pseudohuery/argus-{hub,agent,web}` (mesmo conteúdo, `linux/amd64` e `linux/arm64`).

## Antes de atualizar

1. Leia a seção da versão no [CHANGELOG.md](../CHANGELOG.md). Em mudança **major**, procure por *Breaking* e ajuste `.env` antes de subir.
2. Confira se o `.env` tem as variáveis novas listadas em [.env.example](../.env.example).
3. O hub guarda os dados no Postgres indicado por `ARGUS_STORE_DSN`. `docker compose down` não apaga esse banco. Faça um dump com as ferramentas do próprio Postgres antes de uma atualização major.

## Opção A — imagens publicadas (GHCR)

Fixe a versão com `ARGUS_VERSION` (sem o prefixo `v`). Sem ela, o compose usa `latest`. Para puxar do Docker Hub em vez do GHCR, defina `ARGUS_REGISTRY=docker.io/pseudohuery`.

```bash
export ARGUS_VERSION=0.1.1
# export ARGUS_REGISTRY=docker.io/pseudohuery
docker compose -f examples/compose-minimal/docker-compose.published.yml pull
docker compose -f examples/compose-minimal/docker-compose.published.yml up -d --force-recreate
```

Para usar um compose próprio, aponte `image:` para `ghcr.io/harryraddatz/argus-<componente>:${ARGUS_VERSION}` e repita os mesmos comandos.

## Opção B — build local

A partir do clone, na tag desejada:

```bash
git fetch --tags
git checkout v0.1.1
docker compose up -d --build --force-recreate
```

## Por que `--force-recreate`

`docker compose restart` reinicia o container com a configuração antiga: não relê `env_file` nem troca a imagem. Depois de mudar `.env`, tag ou imagem, recrie os containers com `up -d --force-recreate`.

## Checklist pós-deploy

Ajuste host e portas se não usar os defaults (`8080` hub, `3000` painel).

1. Hub online:

   ```bash
   curl -fsS http://localhost:8080/health
   # {"status":"ok"}
   ```

2. Painel responde 200 e faz proxy da API:

   ```bash
   curl -fsS -o /dev/null -w '%{http_code}\n' http://localhost:3000/
   curl -fsS http://localhost:3000/health
   ```

3. Agent reportando (aguarde um ciclo de `ARGUS_COLLECT_INTERVAL`, default 15s):

   ```bash
   curl -fsS 'http://localhost:8080/api/v1/workloads?since=5m' | head -c 500
   curl -fsS http://localhost:8080/api/v1/fleet/status | head -c 500
   ```

   A lista de workloads não pode vir vazia se houver containers em execução no host.

4. Logs sem erro de autenticação no ingest (token divergente entre hub e agent):

   ```bash
   docker compose logs --since 5m argus-agent | grep -i -E 'error|401' || echo ok
   ```

5. No painel, **Dashboard** e **Workloads** mostram dados recentes.

## Rollback

Volte para a versão anterior e recrie:

```bash
export ARGUS_VERSION=0.1.0
docker compose -f examples/compose-minimal/docker-compose.published.yml up -d --force-recreate
```

No build local: `git checkout v0.1.0 && docker compose up -d --build --force-recreate`.

Se a versão nova alterou o schema do banco, restaure o dump feito antes do deploy.
