# Contributing to Argus

Thanks for contributing. This repository is the public Argus observability library.

Portuguese: [CONTRIBUTING.pt-BR.md](CONTRIBUTING.pt-BR.md).

## Before opening a PR

1. Open an issue with a template (bug, feature, docs, or task). Blank issues are off. Vulnerability: [SECURITY.md](SECURITY.md), not a public issue.
2. Fork + branch from `main`.
3. Keep the diff focused — skip unrelated refactors.
4. Do not include secrets, tokens, or private infra (hosts, internal stacks, production domains).

## Development

```bash
cp .env.example .env
docker compose up -d --build   # full stack
go vet ./...
go test ./...                  # backend (Go 1.27+)
cd web && npm ci && npm run lint && npm test -- --run && npm run build   # panel (Node 24 LTS)
bash .github/scripts/check-no-vps-leak.sh   # optional, local
```

## CI and branch protection

The [`.github/workflows/ci.yml`](.github/workflows/ci.yml) workflow runs on every PR and push to `main`. Jobs run in parallel; a new PR cancels the previous run of the same PR.

| Job | What it checks |
|---|---|
| `changes` | Which components changed in the diff ([`ci-changes.sh`](.github/scripts/ci-changes.sh)) |
| `anti-leak` | Private-infra patterns in the diff |
| `gitleaks` | Committed secrets: PR commit range, or full history on the weekly schedule |
| `vuln-go` | `govulncheck ./...` (reachable Go CVEs only) |
| `vuln-web` | `npm audit --omit=dev --audit-level=high` in `web/` |
| `go` | `go vet`, `go test`, `go build ./cmd/...` — only if hub or agent changed |
| `web` | `npm ci`, `npm run lint`, `npm test -- --run`, `npm run build` — only if web changed |
| `docker (hub/agent/web)` | Build of the changed component's Dockerfile (`load: true`), Trivy `CRITICAL,HIGH` (ignore unfixed), SARIF to Security |
| `smoke` | Compose stack with those images plus a JSON-log sidecar; post-deploy checklist |
| `test` | Aggregator: passes if no job failed (skipped jobs count as ok) |

Components by path:

| Component | Paths |
|---|---|
| hub | `cmd/hub/`, `internal/` (except `internal/agent/`), `go.mod`, `go.sum`, `Dockerfile.hub` |
| agent | `cmd/agent/`, `internal/agent/`, `internal/model/`, `go.mod`, `go.sum`, `Dockerfile.agent` |
| web | `web/` |

A change in `ci.yml`, `ci-changes.sh`, or the smoke compose files, a tag, `workflow_dispatch`, schedule, or a new branch without a base runs everything.

The same workflow also runs **weekly** (Monday) so a CVE with no code change still fails in Actions. [CodeQL](.github/workflows/codeql.yml) (`go` and `javascript-typescript`) runs on every PR and on a weekly schedule; alerts land in **Security → Code scanning**.

The Pages workflow [`.github/workflows/pages.yml`](.github/workflows/pages.yml) runs the **build site** check on every PR (`mkdocs build --strict`). Push to `main` still deploys only when docs paths change.

**Branch protection on `main`:** require a pull request, require the **test** and **build site** checks, require the branch to be up to date, and block force-push. Without those rules, merges can ignore CI.

**Stacked PRs:** if `main` moved, or the PR was opened against a feature branch, retarget the base to `main` and rebase onto `main` before merge. Do not merge a PR whose base is another feature branch.

Dependabot ([`.github/dependabot.yml`](.github/dependabot.yml)) opens grouped weekly PRs for GitHub Actions, Go modules, npm (`web/`), and the web image. CI runs on those PRs like any other.

### Dependency and image alerts

When `vuln-go`, `vuln-web`, or Trivy fails:

1. Upgrade to a patched version (`go get`, `npm update`, or a newer base image) and re-run CI.
2. If there is no patch, open an issue with the CVE, why the risk is accepted, and when you will revisit. Do not add a silent ignore without a linked issue.
3. For npm, use `overrides` only in a PR that cites that issue. `govulncheck` already ignores unreachable findings; do not disable the job to hide a reachable one.

## Release (maintainers)

### Choose the version

Follow [semver](https://semver.org/). While the version is `0.x`, breaking changes bump the **minor**.

| Change | Bump | Example |
|---|---|---|
| Fix without changing the contract | patch | `0.1.0` → `0.1.1` |
| New feature, route, or optional env | minor | `0.1.1` → `0.2.0` |
| Removed/renamed route, incompatible payload, new required env, incompatible database schema | major (minor on `0.x`) | `1.4.2` → `2.0.0` |

Breaking changes go in the CHANGELOG under `### Breaking`, with the migration step for operators.

### Publish

1. Update `CHANGELOG.md` (`[Unreleased]` → `[X.Y.Z] - date`) and the version link in the footer
2. Commit and push to `main`
3. Tag semver and push:

```bash
git tag v0.1.1
git push origin v0.1.1
```

The [`.github/workflows/release.yml`](.github/workflows/release.yml) workflow runs in stages:

1. `verify tag` — semver tag, commit present on `main`, and `[X.Y.Z]` section in the CHANGELOG. Failure here publishes nothing.
2. `ci` — the same PR CI, reused (including smoke).
3. `image (hub/agent/web)` — Trivy on an amd64 image (`CRITICAL,HIGH`, ignore unfixed); then multi-arch build (`linux/amd64`, `linux/arm64`) and parallel push to GHCR and Docker Hub, with SBOM and provenance. A fixable critical/high CVE blocks the push. Tags `X.Y.Z`, `X.Y`, `X` (from `1.0`), `vX.Y.Z`, and `latest`.
4. `docker hub description` — syncs README and short description of each repository from `.github/dockerhub/`.
5. `github release` — notes extracted from the CHANGELOG.

Suffixed tags (`v0.2.0-rc.1`) become pre-releases and do not move `latest`.

Docker Hub uses secrets `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` (token with Read, Write, Delete) and repository variable `DOCKERHUB_NAMESPACE`. Without the variable (for example, on forks), the release publishes only to GHCR.

Post-release smoke test:

```bash
docker compose -f examples/compose-minimal/docker-compose.published.yml pull
docker compose -f examples/compose-minimal/docker-compose.published.yml up -d
curl -s http://localhost:8080/health
```

### Checklist

- [ ] Green CI on `main` at the tag commit
- [ ] `[X.Y.Z]` section in the CHANGELOG with date and items moved from `[Unreleased]`
- [ ] `### Breaking` filled when there is an incompatible change
- [ ] New variables in `.env.example` and `docs/api/configuration.md`
- [ ] Release workflow finished and all three images have tag `X.Y.Z` on GHCR and Docker Hub
- [ ] Smoke test above with `ARGUS_VERSION=X.Y.Z`

Upgrade an existing instance: [docs/deploy.md](docs/deploy.md).

## Documentation site

The site at [harryraddatz.github.io/argus-observability](https://harryraddatz.github.io/argus-observability/) is generated by [`.github/workflows/pages.yml`](.github/workflows/pages.yml) with MkDocs from `docs/` (home is `docs/index.md`), `CHANGELOG.md`, and `CONTRIBUTING.md`. English is the default locale; Portuguese uses the `*.pt.md` suffix.

- A PR that touches these files runs `mkdocs build --strict`: a broken link or page fails the check.
- Merge to `main` and each completed release publish the site.
- A new page in `docs/` must enter the `nav` in `mkdocs.yml`.

Local preview:

```bash
pip install -r .github/pages/requirements.txt
mkdocs serve
```

## Commits

Use short conventional messages in **English** (`feat:`, `fix:`, `docs:`). Same for PR titles, PR bodies, and new CHANGELOG entries. Reference issues in the body when applicable (`Closes #123`). Canonical docs and API identifiers are English; Portuguese files are translations only.

## Pull requests

GitHub fills [`.github/pull_request_template.md`](.github/pull_request_template.md) (Why, What changes, How to validate, checklist). Link the issue (`Closes #123`). CI must pass. If you change API routes, contracts, or UI, update `docs/api/` and `docs/map.md`.

## Code

- Go: follow the existing style in the package you touch.
- Web: React + shadcn/ui in `web/` — reusable components in `web/src/components/`.
- Tests: add or adjust `_test.go` for changed domain logic.

## Conduct

Participants follow [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
