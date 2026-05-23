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

Without --module, nested modules are discovered by scripts/list-nested-modules.sh.
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
      gosec ./...
    fi
  )
  echo "::endgroup::"
}

if [[ -n "$module" ]]; then
  check_module "$module"
  exit 0
fi

"$(dirname "${BASH_SOURCE[0]}")/list-nested-modules.sh" |
  while IFS= read -r dir; do
    check_module "$dir"
  done
