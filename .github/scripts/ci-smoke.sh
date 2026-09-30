#!/usr/bin/env bash
# Post-deploy checklist from docs/deploy.md against a local compose stack.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

COMPOSE=(docker compose -f .github/ci/docker-compose.smoke.yml)
HUB=http://127.0.0.1:8080
PANEL=http://127.0.0.1:3000
LOG_FILE="${SMOKE_LOG_FILE:-smoke-compose-logs.txt}"

dump_logs() {
  "${COMPOSE[@]}" logs --no-color >"$LOG_FILE" 2>&1 || true
}

fail() {
  echo "smoke: $*" >&2
  dump_logs
  exit 1
}

wait_http() {
  local url="$1" tries="${2:-30}"
  local i
  for i in $(seq 1 "$tries"); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done
  fail "timed out waiting for $url"
}

json_ok() {
  python3 -c 'import json,sys; json.load(sys.stdin)'
}

"${COMPOSE[@]}" up -d

wait_http "$HUB/health" 30
curl -fsS "$HUB/health" | json_ok || fail "hub /health is not JSON"

code="$(curl -fsS -o /dev/null -w '%{http_code}' "$PANEL/")"
[[ "$code" == "200" ]] || fail "panel returned HTTP $code"
curl -fsS "$PANEL/health" | json_ok || fail "panel /health is not JSON"

workloads_ok=false
for i in $(seq 1 12); do
  body="$(curl -fsS "$HUB/api/v1/workloads?since=5m" || true)"
  if printf '%s' "$body" | python3 -c 'import json,sys; d=json.load(sys.stdin); sys.exit(0 if isinstance(d,list) and len(d)>0 else 1)'; then
    workloads_ok=true
    break
  fi
  sleep 5
done
[[ "$workloads_ok" == "true" ]] || fail "workloads stayed empty"

curl -fsS "$HUB/api/v1/fleet/status" | json_ok || fail "fleet/status is not JSON"

agent_logs="$("${COMPOSE[@]}" logs --no-color argus-agent || true)"
if printf '%s' "$agent_logs" | grep -qiE '401|unauthorized'; then
  fail "agent logs contain ingest auth errors"
fi

echo "smoke: ok"
