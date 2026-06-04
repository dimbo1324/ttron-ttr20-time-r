#!/usr/bin/env sh
set -eu

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required for check-go-format.sh" >&2
  exit 1
fi

python3 - <<'PY'
from pathlib import Path
import subprocess
import sys

root = Path.cwd()
result = subprocess.run(
    ["git", "ls-files", "*.go"],
    cwd=root,
    check=True,
    capture_output=True,
    text=True,
)

failures = []
for rel in result.stdout.splitlines():
    if not rel or rel.startswith("legacy/") or rel.startswith("web/node_modules/"):
        continue
    path = root / rel
    formatted = subprocess.run(
        ["gofmt", str(path)],
        cwd=root,
        check=True,
        capture_output=True,
    ).stdout.decode("utf-8")
    original = path.read_text(encoding="utf-8")
    if original.replace("\r\n", "\n").replace("\r", "\n") != formatted.replace("\r\n", "\n").replace("\r", "\n"):
        failures.append(rel)

if failures:
    print("Go files need gofmt:", file=sys.stderr)
    print("\n".join(failures), file=sys.stderr)
    sys.exit(1)

print("go format check passed")
PY
