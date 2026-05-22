#!/usr/bin/env bash
set -euo pipefail

workflow=".github/workflows/ci.yml"

if [[ ! -f "$workflow" ]]; then
  echo "$workflow does not exist" >&2
  exit 2
fi

expected="$(mktemp)"
actual="$(mktemp)"
trap 'rm -f "$expected" "$actual"' EXIT

find . -name go.sum -printf '%P\n' | sort >"$expected"

awk '
  /cache-dependency-path: \|/ { in_block = 1; next }
  in_block && /^            / {
    sub(/^            /, "")
    print
    next
  }
  in_block { in_block = 0 }
' "$workflow" | sort -u >"$actual"

if ! diff -u "$expected" "$actual"; then
  echo "CI cache-dependency-path entries do not match repository go.sum files" >&2
  exit 1
fi
