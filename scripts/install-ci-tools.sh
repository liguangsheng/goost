#!/usr/bin/env bash
set -euo pipefail

profile=""

usage() {
  cat <<'USAGE'
Usage: scripts/install-ci-tools.sh --root|--split

Options:
  --root   Install tools used by the root module CI gate.
  --split  Install tools used by the nested module CI gate.
USAGE
}

while (($#)); do
  case "$1" in
    --root)
      if [[ -n "$profile" ]]; then
        echo "only one profile may be selected" >&2
        usage >&2
        exit 2
      fi
      profile="root"
      shift
      ;;
    --split)
      if [[ -n "$profile" ]]; then
        echo "only one profile may be selected" >&2
        usage >&2
        exit 2
      fi
      profile="split"
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

if [[ -z "$profile" ]]; then
  echo "missing profile: use --root or --split" >&2
  usage >&2
  exit 2
fi

: "${STATICCHECK_VERSION:=latest}"
: "${GOVULNCHECK_VERSION:=latest}"
: "${GOSEC_VERSION:=latest}"

if [[ "$profile" == "root" ]]; then
  : "${GOLANGCI_LINT_VERSION:=latest}"
  go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}"
fi

go install "honnef.co/go/tools/cmd/staticcheck@${STATICCHECK_VERSION}"
go install "golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}"
go install "github.com/securego/gosec/v2/cmd/gosec@${GOSEC_VERSION}"
