#!/usr/bin/env bash
set -euo pipefail

mode="full"
dir="."

usage() {
  cat <<'USAGE'
Usage: scripts/check-root.sh [--quick|--full] [--module DIR]

Options:
  --quick       Run tidy, vet, ordinary tests, and golangci-lint.
  --full        Run the full root gate, including race tests and security checks.
  --module DIR  Run the selected gate from DIR instead of the repository root.
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
      dir="$2"
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

cd "$dir"

go mod tidy -diff
go vet ./...

if [[ "$mode" == "quick" ]]; then
  go test ./...
  golangci-lint run ./...
  exit 0
fi

go test -race -coverprofile=coverage.out -covermode=atomic ./...
golangci-lint run ./...
staticcheck ./...
govulncheck ./...
gosec -exclude=G115,G301,G302,G304 -exclude-dir=examples ./...
