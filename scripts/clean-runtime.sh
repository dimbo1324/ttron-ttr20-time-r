#!/usr/bin/env sh
set -euo pipefail

dry_run=0
if [ "${1:-}" = "--dry-run" ]; then
  dry_run=1
elif [ "${1:-}" != "" ]; then
  echo "usage: scripts/clean-runtime.sh [--dry-run]" >&2
  exit 2
fi

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

remove_path() {
  path="$1"
  [ -e "$path" ] || return 0
  case "$path" in
    "$root"/*) ;;
    *) echo "refusing to clean path outside repository: $path" >&2; exit 1 ;;
  esac
  rel="${path#"$root"/}"
  if [ "$dry_run" -eq 1 ]; then
    echo "would remove $rel"
  else
    echo "remove $rel"
    rm -rf -- "$path"
  fi
}

for rel in \
  tmp \
  runtime \
  bin \
  dist \
  coverage.out \
  web/dist \
  web/.vite \
  web/tsconfig.app.tsbuildinfo
do
  remove_path "$root/$rel"
done

find "$root" -maxdepth 1 -type f \( -name "*.log" -o -name "*.out" -o -name "*.tsbuildinfo" \) -print | while IFS= read -r path; do
  remove_path "$path"
done

if [ "$dry_run" -eq 1 ]; then
  echo "cleanup dry-run complete"
else
  echo "cleanup complete"
fi
