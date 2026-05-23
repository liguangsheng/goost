#!/usr/bin/env bash
set -euo pipefail

mode="quick"
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat <<'USAGE'
Usage: scripts/check-stress.sh [--quick|--race]

Options:
  --quick  Run ordinary stress-focused package tests.
  --race   Run the same stress-focused packages under the race detector.

Long-running ad hoc stress loops should stay outside this script until they are
stable enough for repeatable local use.
USAGE
}

while (($#)); do
  case "$1" in
    --quick)
      mode="quick"
      shift
      ;;
    --race)
      mode="race"
      shift
      ;;
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
done

packages=(
  ./batcher
  ./fanout
  ./keyedmutex
  ./pool
  ./ttlmap
)

cd "$repo_root"

if [[ "$mode" == "quick" ]]; then
  go test -run 'Stress' "${packages[@]}"
  exit 0
fi

go test -race -run 'Stress' "${packages[@]}"
