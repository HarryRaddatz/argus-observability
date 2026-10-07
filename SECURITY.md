# Security policy

## Supported versions

Security fixes land in the **latest published release** on [Releases](https://github.com/HarryRaddatz/argus-observability/releases). While the project is on `0.x`, use the newest tag (`v0.3.x` at the moment).

| Line | Receives a fix |
|---|---|
| Latest tag `vX.Y.Z` | Yes |
| Older tags | No — upgrade using the [deploy runbook](docs/deploy.md) |

Portuguese: [SECURITY.pt-BR.md](SECURITY.pt-BR.md).

## How to report

Do not open a public issue or a PR with an exploit PoC.

1. Open a [private advisory](https://github.com/HarryRaddatz/argus-observability/security/advisories/new) (Private vulnerability reporting).
2. Include component (`hub`, `agent`, `web`), version or commit, impact, and steps to reproduce.
3. Do not paste tokens, production dumps, or `.env` files.

Expected response within **7 days**. The fix is published in `CHANGELOG.md` under `Security`.

When `ARGUS_AGENT_TOKEN` is set, query routes require the same Bearer token as ingest. `/health` stays open. Do not expose the hub port to the internet; the panel proxy is what presents the token to the hub. See [configuration](docs/api/configuration.md).
