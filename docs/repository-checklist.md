# Repository Checklist

Use this checklist before public demos or GitHub releases.

## Source Hygiene

- `git status` is clean.
- No `bin/`, `dist/`, `tmp/`, logs, or local `.exe` files are tracked.
- `.gitignore` and `.dockerignore` cover runtime artifacts.

## Checks

- `go fmt ./...`
- `go test ./...`
- `go build ./...`
- `.\scripts\check-architecture.ps1`
- `docker compose config`
- `docker compose build`
- `.\scripts\check-doc-links.ps1`
- `.\scripts\release-check.ps1`

## Smoke

- Docker Compose stack starts.
- `GET /health` returns `200`.
- `GET /api/v1/ready` returns `200`.
- `GET /api/v1/overview` returns `200`.
- `GET /api/v1/events` returns `200`.
- `GET /api/v1/export/events.json` returns `200`.
- `GET /api/v1/export/events.csv` returns `200`.
- `GET /api/v1/export/overview.json` returns `200`.
- `GET /metrics` returns `200`.

## Documentation

- README quick start works from a fresh clone.
- Docs index links all major docs.
- Troubleshooting covers known local issues.
- Release notes/changelog are current.

## Safety

- README and docs state no auth/TLS/persistence yet.
- No secrets are committed.
- No production-readiness claims are made beyond the implemented local baseline.
