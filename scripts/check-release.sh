#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat <<'USAGE'
Usage: scripts/check-release.sh

Runs the local pre-release gate:
  1. CHANGELOG format check
  2. doc link check
  3. scripts self-check and CI cache path alignment
  4. root full gate
  5. split-module full gate
USAGE
}

if (($#)); then
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
fi

cd "$repo_root"

echo "::group::changelog format"
for f in CHANGELOG.md CHANGELOG.zh.md; do
  if [[ ! -f "$f" ]]; then
    echo "missing $f" >&2; exit 1
  fi
  if ! grep -q '## \[Unreleased\]\|## \[v[0-9]' "$f"; then
    echo "$f: missing release heading (## [Unreleased] or ## [vX.Y.Z])" >&2; exit 1
  fi
  if ! grep -q '^### ' "$f"; then
    echo "$f: missing section heading (### Added / ### Changed / etc.)" >&2; exit 1
  fi
done
echo "changelog format ok"
echo "::endgroup::"

./scripts/check-doc-links.sh
./scripts/check-scripts.sh
./scripts/check-root.sh --full
./scripts/check-split-modules.sh --full
