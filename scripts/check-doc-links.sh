#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
broken=0

usage() {
  cat <<'USAGE'
Usage: scripts/check-doc-links.sh

Check relative markdown links in all .md files. Skips external URLs,
mailto links, and anchor-only links.
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

while IFS= read -r -d '' file; do
  dir="$(dirname "$file")"
  lineno=0
  while IFS= read -r line; do
    ((lineno++)) || true
    # Extract link targets: [text](target)
    while IFS= read -r target; do
      case "$target" in
        ""|"#"*|http://*|https://*|mailto:*) continue ;;
      esac
      # Strip anchor
      target="${target%%#*}"
      [[ -z "$target" ]] && continue
      # Resolve relative to file directory
      resolved="$dir/$target"
      if [[ ! -e "$resolved" ]]; then
        echo "$file:$lineno: broken link: $target" >&2
        broken=1
      fi
    done < <(grep -oP '(?<=\]\()[^)]+' "$line" 2>/dev/null || true)
  done < "$file"
done < <(find . -name "*.md" -not -path "./.git/*" -not -path "./.agents/*" -print0 | sort -z)

if ((broken)); then
  echo "broken links found" >&2
  exit 1
fi
echo "all doc links ok"
