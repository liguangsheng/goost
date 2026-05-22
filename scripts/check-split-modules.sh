#!/usr/bin/env bash
set -euo pipefail

mode="full"
module=""

usage() {
  cat <<'USAGE'
Usage: scripts/check-split-modules.sh [--quick|--full] [--module DIR]

Options:
  --quick       Run tidy, vet, ordinary tests, and staticcheck.
  --full        Run the full split-module gate, including vulnerability and security checks.
  --module DIR  Check only the nested module at DIR.
USAGE
}

while (($#)); do
  case "$1" in
    --quick)
      mode="quick"
      shift
      ;;
    --full)
      mode="full"
      shift
      ;;
    --module)
      if (($# < 2)); then
        echo "missing DIR for --module" >&2
        usage >&2
        exit 2
      fi
      module="$2"
      shift 2
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

if [[ "$mode" != "quick" && "$mode" != "full" ]]; then
  echo "unsupported mode: $mode" >&2
  exit 2
fi

check_module() {
  local dir="$1"
  if [[ ! -f "$dir/go.mod" ]]; then
    echo "$dir does not contain go.mod" >&2
    exit 2
  fi

  echo "::group::$dir"
  (
    cd "$dir"
    go mod tidy -diff
    go vet ./...
    go test ./...
    staticcheck ./...
    if [[ "$mode" == "full" ]]; then
      govulncheck ./...
      gosec -exclude=G115,G301,G302,G304 ./...
    fi
  )
  echo "::endgroup::"
}

if [[ -n "$module" ]]; then
  check_module "$module"
  exit 0
fi

find . -mindepth 2 -name go.mod -print0 |
  while IFS= read -r -d '' mod; do
    dir="$(dirname "$mod")"
    check_module "$dir"
  done
