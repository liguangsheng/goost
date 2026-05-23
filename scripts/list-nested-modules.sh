#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$repo_root"

find . \
  -path './.git' -prune -o \
  -path './.agents' -prune -o \
  -name go.mod \
  -print |
  while IFS= read -r mod; do
    dir="${mod%/go.mod}"
    if [[ "$dir" != "." ]]; then
      printf '%s\n' "$dir"
    fi
  done |
  sort
