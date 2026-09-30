# Contributing to Argus

Obrigado por contribuir. Este repositório é a biblioteca pública de observabilidade Argus.

[English](CONTRIBUTING.md)

## Antes de abrir PR

1. Abra uma issue pelo template (bug, feature, docs ou task). Issues em branco estão desligadas. Vulnerabilidade: [SECURITY.md](SECURITY.md), não issue pública.
2. Fork + branch a partir de `main`.
3. Mantenha o diff focado — evite refactors não relacionados.
4. Não inclua secrets, tokens ou referências a infra privada (hosts, stacks internas, domínios de produção).

## Desenvolvimento

```bash
cp .env.example .env
docker compose up -d --build   # stack completa
go vet ./...
go test ./...                  # backend (Go 1.27+)
cd web && npm ci && npm run lint && npm run build   # painel (Node 24 LTS)
bash .github/scripts/check-no-vps-leak.sh   # opcional, local
```

## CI e branch protection

O workflow [`.github/workflows/ci.yml`](.github/workflows/ci.yml) roda em todo PR e push para `main`. Os jobs rodam em paralelo; um PR novo cancela a execução anterior do mesmo PR.

| Job | O que valida |
|---|---|
| `changes` | Quais componentes mudaram no diff ([`ci-changes.sh`](.github/scripts/ci-changes.sh)) |
| `anti-leak` | Padrões de infra privada no diff |
| `go` | `go vet`, `go test`, `go build ./cmd/...` — só se hub ou agent mudou |
| `web` | `npm ci`, `npm run lint`, `npm run build` — só se web mudou |
| `docker (hub/agent/web)` | Build do Dockerfile do componente que mudou, sem push, com cache compartilhado com o release |
| `test` | Agregador: passa se nenhum job falhou (jobs pulados contam como ok) |

Componentes por path:

| Componente | Paths |
|---|---|
| hub | `cmd/hub/`, `internal/` (exceto `internal/agent/`), `go.mod`, `go.sum`, `Dockerfile.hub` |
| agent | `cmd/agent/`, `internal/agent/`, `internal/model/`, `go.mod`, `go.sum`, `Dockerfile.agent` |
| web | `web/` |

Mudança em `ci.yml` ou `ci-changes.sh`, tag, `workflow_dispatch` ou branch nova sem base roda tudo.

O workflow Pages [`.github/workflows/pages.yml`](.github/workflows/pages.yml) roda o check **build site** em todo PR (`mkdocs build --strict`). Push em `main` só faz deploy quando paths de docs mudam.

**Branch protection em `main`:** exigir pull request, os checks **test** e **build site**, branch atualizada, e bloquear force-push. Sem essas regras, merges podem ignorar o CI.

**PRs empilhados:** se `main` andou, ou o PR foi aberto contra uma feature, redirecione a base para `main` e faça rebase em `main` antes do merge. Não faça merge de PR cuja base é outra feature.

O Dependabot ([`.github/dependabot.yml`](.github/dependabot.yml)) abre PRs agrupados semanais para GitHub Actions, módulos Go, npm (`web/`) e a imagem web. O CI roda nesses PRs como nos demais.

## Release (maintainers)

### Escolher a versão

Siga [semver](https://semver.org/). Enquanto a versão for `0.x`, mudanças incompatíveis sobem o **minor**.

| Mudança | Bump | Exemplo |
|---|---|---|
| Correção sem alterar contrato | patch | `0.1.0` → `0.1.1` |
| Funcionalidade nova, rota ou env opcional | minor | `0.1.1` → `0.2.0` |
| Rota removida/renomeada, payload incompatível, env obrigatória nova, schema SQLite incompatível | major (minor em `0.x`) | `1.4.2` → `2.0.0` |

Mudanças incompatíveis entram no CHANGELOG sob `### Breaking`, com o passo de migração para quem atualiza.

### Publicar

1. Atualize `CHANGELOG.md` (seção `[Unreleased]` → `[X.Y.Z] - data`) e o link da versão no rodapé
2. Commit e push em `main`
3. Tag semver e push:

```bash
git tag v0.1.1
git push origin v0.1.1
```

O workflow [`.github/workflows/release.yml`](.github/workflows/release.yml) roda em etapas:

1. `verify tag` — tag semver, commit presente em `main` e seção `[X.Y.Z]` no CHANGELOG. Falha aqui não publica nada.
2. `ci` — o mesmo CI de PR, reutilizado.
3. `image (hub/agent/web)` — build multi-arch (`linux/amd64`, `linux/arm64`) e push em paralelo para GHCR e Docker Hub, com SBOM e provenance. Tags `X.Y.Z`, `X.Y`, `X` (a partir de `1.0`), `vX.Y.Z` e `latest`.
4. `docker hub description` — sincroniza README e descrição curta de cada repositório a partir de `.github/dockerhub/`.
5. `github release` — notas extraídas do CHANGELOG.

Tags com sufixo (`v0.2.0-rc.1`) viram pre-release e não movem `latest`.

O Docker Hub usa os secrets `DOCKERHUB_USERNAME` e `DOCKERHUB_TOKEN` (token com permissão Read, Write, Delete) e a variável de repositório `DOCKERHUB_NAMESPACE`. Sem a variável (por exemplo, em forks), o release publica só no GHCR.

Smoke test pós-release:

```bash
docker compose -f examples/compose-minimal/docker-compose.published.yml pull
docker compose -f examples/compose-minimal/docker-compose.published.yml up -d
curl -s http://localhost:8080/health
```

### Checklist

- [ ] CI verde em `main` no commit da tag
- [ ] Seção `[X.Y.Z]` no CHANGELOG com data e itens de `[Unreleased]` movidos
- [ ] `### Breaking` preenchido quando houver mudança incompatível
- [ ] Novas variáveis em `.env.example` e `docs/api/configuration.md`
- [ ] Workflow Release concluído e as três imagens com a tag `X.Y.Z` no GHCR e no Docker Hub
- [ ] Smoke test acima com `ARGUS_VERSION=X.Y.Z`

Atualizar uma instância existente: [docs/deploy.md](docs/deploy.md).

## Site de documentação

O site em [harryraddatz.github.io/argus-observability](https://harryraddatz.github.io/argus-observability/) é gerado pelo workflow [`.github/workflows/pages.yml`](.github/workflows/pages.yml) com MkDocs a partir de `docs/` (início em `docs/index.md`), `CHANGELOG.md` e `CONTRIBUTING.md`. Inglês é o idioma padrão; português usa o sufixo `*.pt.md`.

- PR que toca esses arquivos roda `mkdocs build --strict`: link ou página quebrada falha o check.
- Merge em `main` e cada release concluída publicam o site.
- Página nova em `docs/` precisa entrar no `nav` do `mkdocs.yml`.

Preview local:

```bash
pip install -r .github/pages/requirements.txt
mkdocs serve
```

## Commits

Use mensagens convencionais curtas **em inglês** (`feat:`, `fix:`, `docs:`). O mesmo vale para título e corpo de PR e para entradas novas do CHANGELOG. Referencie issues no corpo quando aplicável (`Closes #123`). Docs canônicos e identificadores de API são em inglês; os arquivos em português são só tradução.

## Pull requests

O GitHub preenche [`.github/pull_request_template.md`](.github/pull_request_template.md) (Why, What changes, How to validate, checklist). Vincule a issue (`Closes #123`). CI deve passar. Se alterar rotas API, contratos ou UI, atualize `docs/api/` e `docs/map.md`.

## Código

- Go: siga o estilo existente no pacote tocado.
- Web: React + shadcn/ui em `web/` — componentes reutilizáveis em `web/src/components/`.
- Testes: adicione ou ajuste `_test.go` para lógica de domínio alterada.

## Conduta

Participantes seguem o [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
