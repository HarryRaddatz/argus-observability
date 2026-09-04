#!/usr/bin/env bash
# Print CHANGELOG section for a semver tag (e.g. 0.1.1 or v0.1.1).
set -euo pipefail

VERSION="${1#v}"
FILE="${2:-CHANGELOG.md}"

if [[ ! -f "$FILE" ]]; then
  echo "Missing $FILE" >&2
  exit 1
fi

awk -v ver="$VERSION" '
  $0 ~ "^## \\[" ver "\\]" { found=1; print; next }
  found && /^## \[/ { exit }
  found { print }
' "$FILE"
