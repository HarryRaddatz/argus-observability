#!/usr/bin/env bash
# Fail CI when versioned files contain infra-private patterns (see .cursor/rules/no-vps-leak.mdc).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

EXCLUDE=(
  '.cursor/rules/no-vps-leak.mdc'
  '.github/scripts/check-no-vps-leak.sh'
)

should_exclude() {
  local f="$1"
  for ex in "${EXCLUDE[@]}"; do
    [[ "$f" == "$ex" ]] && return 0
  done
  return 1
}

if [[ -n "${BASE_REF:-}" ]] && git rev-parse "$BASE_REF" >/dev/null 2>&1; then
  mapfile -t FILES < <(git diff --name-only --diff-filter=ACMRT "$BASE_REF" HEAD)
  echo "Anti-leak: checking ${#FILES[@]} changed file(s) vs ${BASE_REF}"
else
  mapfile -t FILES < <(git ls-files)
  echo "Anti-leak: checking ${#FILES[@]} tracked file(s)"
fi

PATTERNS=(
  'venuz'
  'hostinger'
  '/root/services'
  '/root/infrastructure'
  'npmplus-shared'
  'village-postgres'
  'agendamentoapi'
  '127\.0\.0\.1:8088'
  '127\.0\.0\.1:5173'
)

failed=0
for file in "${FILES[@]}"; do
  [[ -z "$file" ]] && continue
  should_exclude "$file" && continue
  [[ -f "$file" ]] || continue
  if file "$file" | grep -qE 'binary|executable'; then
    continue
  fi
  for pat in "${PATTERNS[@]}"; do
    if grep -qiE "$pat" "$file" 2>/dev/null; then
      echo "FAIL: $file matches forbidden pattern: $pat"
      grep -niE "$pat" "$file" | head -3
      failed=1
    fi
  done
done

if [[ "$failed" -ne 0 ]]; then
  echo "Anti-leak check failed. See .cursor/rules/no-vps-leak.mdc"
  exit 1
fi

echo "Anti-leak check passed."
