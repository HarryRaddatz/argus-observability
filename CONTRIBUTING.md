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
go test ./...                  # backend (Go 1.22+)
cd web && npm ci && npm run build
```

## Commits

Use mensagens convencionais curtas (`feat:`, `fix:`, `docs:`). Referencie issues no corpo quando aplicável (`Closes #123`).

## Pull requests

- Descreva o **porquê** e como validar.
- CI deve passar (Go tests + build web).
- Docs: se alterar rotas API, contratos ou UI, atualize `docs/api/` e `docs/map.md`.

## Código

- Go: siga o estilo existente no pacote tocado.
- Web: React + shadcn/ui em `web/` — componentes reutilizáveis em `web/src/components/`.
- Testes: adicione ou ajuste `_test.go` para lógica de domínio alterada.

## Conduta

Participantes seguem o [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
