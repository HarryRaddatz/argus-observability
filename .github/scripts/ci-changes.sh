#!/usr/bin/env bash
# Decide which CI components run (hub, agent, web) from files changed vs BASE_REF.
# Without a usable BASE_REF (tag, workflow_dispatch, new branch) or when CI itself
# changes, every component runs.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

COMPONENTS=(hub agent web)

FULL_RUN_FILES=(
  '.github/workflows/ci.yml'
  '.github/scripts/ci-changes.sh'
  '.github/scripts/ci-smoke.sh'
  '.github/ci/docker-compose.smoke.yml'
)

matches() {
  local component="$1" f="$2"
  case "$component" in
    hub)
      [[ "$f" =~ ^(cmd/hub/|go\.mod$|go\.sum$|Dockerfile\.hub$) ]] && return 0
      [[ "$f" == internal/* && "$f" != internal/agent/* ]]
      ;;
    agent)
      [[ "$f" =~ ^(cmd/agent/|internal/agent/|internal/model/|go\.mod$|go\.sum$|Dockerfile\.agent$) ]]
      ;;
    web)
      [[ "$f" == web/* ]]
      ;;
  esac
}

full=false
if [[ "${FORCE_ALL:-false}" == "true" ]]; then
  full=true
  echo "CI changes: full run requested"
elif [[ -z "${BASE_REF:-}" ]] || ! git rev-parse --verify -q "${BASE_REF}^{commit}" >/dev/null; then
  full=true
  echo "CI changes: no usable base ref, running everything"
else
  mapfile -t FILES < <(git diff --name-only "$BASE_REF" HEAD)
  echo "CI changes: ${#FILES[@]} changed file(s) vs ${BASE_REF}"
  for f in "${FILES[@]}"; do
    for ff in "${FULL_RUN_FILES[@]}"; do
      [[ "$f" == "$ff" ]] && full=true
    done
  done
  [[ "$full" == "true" ]] && echo "CI changes: CI definition changed, running everything"
fi

declare -A RESULT
for c in "${COMPONENTS[@]}"; do
  RESULT[$c]=false
  if [[ "$full" == "true" ]]; then
    RESULT[$c]=true
    continue
  fi
  for f in "${FILES[@]}"; do
    if matches "$c" "$f"; then
      RESULT[$c]=true
      break
    fi
  done
done

go=false
[[ "${RESULT[hub]}" == "true" || "${RESULT[agent]}" == "true" ]] && go=true

docker=false
for c in "${COMPONENTS[@]}"; do
  [[ "${RESULT[$c]}" == "true" ]] && docker=true
done

OUT="${GITHUB_OUTPUT:-/dev/stdout}"
{
  for c in "${COMPONENTS[@]}"; do
    echo "$c=${RESULT[$c]}"
  done
  echo "go=$go"
  echo "docker=$docker"
} >> "$OUT"
