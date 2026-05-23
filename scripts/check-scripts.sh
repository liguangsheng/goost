#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$repo_root"

scripts=(
  scripts/check-ci-cache-paths.sh
  scripts/check-release.sh
  scripts/check-root.sh
  scripts/check-scripts.sh
  scripts/check-split-modules.sh
  scripts/check-stress.sh
  scripts/install-ci-tools.sh
  scripts/list-nested-modules.sh
)

for script in "${scripts[@]}"; do
  if [[ ! -x "$script" ]]; then
    echo "$script is not executable" >&2
    exit 1
  fi
  bash -n "$script"
done

./scripts/check-root.sh --help >/dev/null
./scripts/check-split-modules.sh --help >/dev/null
./scripts/check-release.sh --help >/dev/null
./scripts/check-stress.sh --help >/dev/null

./scripts/check-ci-cache-paths.sh

if ! ./scripts/list-nested-modules.sh | grep -qx './examples'; then
  echo "scripts/list-nested-modules.sh did not discover ./examples" >&2
  exit 1
fi

if ./scripts/list-nested-modules.sh | grep -q '^./.agents'; then
  echo "scripts/list-nested-modules.sh must ignore .agents" >&2
  exit 1
fi
