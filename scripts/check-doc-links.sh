#!/usr/bin/env sh
set -eu

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required for check-doc-links.sh" >&2
  exit 1
fi

python3 - <<'PY'
from pathlib import Path
from urllib.parse import unquote
import re
import sys

root = Path.cwd()
files = []
for name in ["README.md", "CONTRIBUTING.md", "SECURITY.md", "CHANGELOG.md"]:
    p = root / name
    if p.exists():
        files.append(p)
for folder in ["docs", "examples", ".github"]:
    p = root / folder
    if p.exists():
        files.extend(p.rglob("*.md"))

pattern = re.compile(r"!?\[[^\]]+\]\(([^)]+)\)")
failures = []
for file in files:
    text = file.read_text(encoding="utf-8")
    for match in pattern.finditer(text):
        target = match.group(1).strip()
        if not target or target.startswith("#"):
            continue
        if re.match(r"^(https?://|mailto:|tel:)", target):
            continue
        if target.startswith("<") and target.endswith(">"):
            target = target[1:-1]
        target = re.split(r"\s+", target)[0].split("#", 1)[0]
        if not target:
            continue
        target = unquote(target)
        candidate = (root / target.lstrip("/")) if target.startswith("/") else (file.parent / target)
        if not candidate.exists():
            failures.append(f"{file.relative_to(root)} -> {match.group(1)}")

if failures:
    print("broken local markdown links:", file=sys.stderr)
    print("\n".join(failures), file=sys.stderr)
    sys.exit(1)

print("doc link check passed")
PY
