#!/usr/bin/env bash
set -euo pipefail

find . -mindepth 2 -name go.mod -print0 |
  while IFS= read -r -d '' mod; do
    dir="$(dirname "$mod")"
    echo "::group::$dir"
    (
      cd "$dir"
      go vet ./...
      go test ./...
      staticcheck ./...
      govulncheck ./...
      gosec -exclude=G115,G301,G302,G304 ./...
    )
    echo "::endgroup::"
  done
