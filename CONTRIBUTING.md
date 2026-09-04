# Contributing to Argus

Obrigado por contribuir. Este repositório é a biblioteca pública de observabilidade Argus.

## Antes de abrir PR

1. Abra uma issue descrevendo a mudança (bug, feature ou doc).
2. Fork + branch a partir de `main`.
3. Mantenha o diff focado — evite refactors não relacionados.
4. Não inclua secrets, tokens ou referências a infra privada (hosts, stacks internas, domínios de produção).

## Desenvolvimento

```bash
cp .env.example .env
docker compose up -d --build   # stack completa
go vet ./...
go test ./...                  # backend (Go 1.22+)
cd web && npm ci && npm run build
bash .github/scripts/check-no-vps-leak.sh   # opcional, local
```

## CI e branch protection

O workflow [`.github/workflows/ci.yml`](.github/workflows/ci.yml) (job `test`) roda em todo PR e push para `main`:

1. Anti-leak — padrões de infra privada no diff
2. `go vet ./...`
3. `go test ./...`
4. `npm ci && npm run build` (web)

**Branch protection (config manual no GitHub):** em *Settings → Branches → main*, marque *Require status checks* e selecione o check **test**. Sem isso, merges podem ignorar o CI.

## Release (maintainers)

1. Atualize `CHANGELOG.md` (seção `[Unreleased]` → `[X.Y.Z] - data`)
2. Commit e push em `main`
3. Tag semver e push:

```bash
git tag v0.1.1
git push origin v0.1.1
```

O workflow [`.github/workflows/release.yml`](.github/workflows/release.yml) publica:

- GitHub Release (notas do CHANGELOG)
- Imagens `ghcr.io/harryraddatz/argus-{hub,agent,web}` com tag semver e `latest`

Smoke test pós-release:

```bash
docker compose -f examples/compose-minimal/docker-compose.published.yml pull
docker compose -f examples/compose-minimal/docker-compose.published.yml up -d
curl -s http://localhost:8080/health
```

## Commits

Use mensagens convencionais curtas (`feat:`, `fix:`, `docs:`). Referencie issues no corpo quando aplicável (`Closes #123`).

## Pull requests

- Descreva o **porquê** e como validar.
- CI deve passar (anti-leak, go vet, go test, build web).
- Docs: se alterar rotas API, contratos ou UI, atualize `docs/api/` e `docs/map.md`.

## Código

- Go: siga o estilo existente no pacote tocado.
- Web: React + shadcn/ui em `web/` — componentes reutilizáveis em `web/src/components/`.
- Testes: adicione ou ajuste `_test.go` para lógica de domínio alterada.

## Conduta

Participantes seguem o [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
