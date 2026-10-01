#!/usr/bin/env bash
# Scan committed secrets: PR range on pull_request, full history on schedule,
# otherwise the pushed commits.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

IMAGE="ghcr.io/gitleaks/gitleaks:v8.24.3"
EVENT="${GITHUB_EVENT_NAME:-}"
SHA="${GITHUB_SHA:-HEAD}"

run_detect() {
  docker run --rm -v "$ROOT:/src" -w /src "$IMAGE" detect \
    --source /src --verbose --redact --exit-code 1 "$@"
}

if [[ "$EVENT" == "pull_request" ]]; then
  [[ -n "${BASE_SHA:-}" ]] || { echo "gitleaks: BASE_SHA is required on pull_request"; exit 1; }
  run_detect --log-opts "${BASE_SHA}..${SHA}"
elif [[ "$EVENT" == "schedule" ]]; then
  run_detect
elif [[ -n "${BEFORE_SHA:-}" && "$BEFORE_SHA" != "0000000000000000000000000000000000000000" ]]; then
  run_detect --log-opts "${BEFORE_SHA}..${SHA}"
else
  run_detect
fi
