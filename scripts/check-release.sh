#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat <<'USAGE'
Usage: scripts/check-release.sh

Runs the local pre-release gate:
  1. scripts self-check and CI cache path alignment
  2. root full gate
  3. split-module full gate
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

./scripts/check-scripts.sh
./scripts/check-root.sh --full
./scripts/check-split-modules.sh --full
