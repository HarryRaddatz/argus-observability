# Política de segurança

[English](SECURITY.md)

## Versões suportadas

Correções de segurança entram na **última release** publicada em [Releases](https://github.com/HarryRaddatz/argus-observability/releases). Enquanto o projeto estiver em `0.x`, use a tag mais recente (`v0.3.x` no momento).

| Linha | Recebe correção |
|---|---|
| Última tag `vX.Y.Z` | Sim |
| Tags anteriores | Não — atualize pela [runbook de deploy](docs/deploy.md) |

## Como reportar

Não abra issue pública nem PR com PoC de exploração.

1. Abra um [aviso privado](https://github.com/HarryRaddatz/argus-observability/security/advisories/new) (Private vulnerability reporting).
2. Inclua componente (`hub`, `agent`, `web`), versão ou commit, impacto e passos para reproduzir.
3. Não cole tokens, dumps de produção nem `.env`.

Resposta esperada em até **7 dias**. Correção publicada no `CHANGELOG.md` na seção `Security`.

Rotas de consulta do hub não têm autenticação. Não exponha a porta do hub à internet sem um proxy autenticado na frente — ver [configuração](docs/api/configuration.md).
