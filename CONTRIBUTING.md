# Contributing

Thanks for taking a look at this project. This repository is structured as a
small industrial protocol simulation platform, so changes should preserve the
separation between protocol core, service runtime, API adapters, Web UI, and
deployment tooling.

## Prerequisites

- Go version declared in `go.mod`.
- Node 22 or a recent LTS release.
- Docker Desktop or compatible Docker daemon for Compose checks.
- PowerShell on Windows, or `sh`/`make` on Unix-like systems.

## Setup

```powershell
git clone https://github.com/dimbo1324/ttron-ttr20-time-r.git
cd ttron-ttr20-time-r
go test ./...
cd web
npm ci
npm run build
```

## Backend Checks

```powershell
go fmt ./...
.\scripts\check-go-format.ps1
go test ./...
go build ./...
.\scripts\check-architecture.ps1
```

## Frontend Checks

```powershell
cd web
npm ci
npm run typecheck
npm run lint
npm run build
```

## Docker Checks

```powershell
docker compose config
docker compose build
docker compose up -d
docker compose down -v
```

## Architecture Rules

- Keep `internal/protocol` independent of TCP, gRPC, HTTP, config, logging, Web,
  Docker, emulator, and gateway packages.
- Do not duplicate protocol business logic in API or UI adapters.
- Do not modify generated gRPC files by hand.
- Do not change wire protocol or gRPC contracts without a dedicated design
  change and tests.

## Generated Artifacts

Do not commit:

- `bin/`
- `dist/`
- `tmp/`
- `runtime/`
- logs
- local `.exe` files
- `web/node_modules/`
- `web/dist/`
- `web/*.tsbuildinfo`

Use cleanup dry-run before committing when local build output has accumulated:

```powershell
.\scripts\clean-runtime.ps1 -DryRun
```

Active Go code lives in `cmd/` and `internal/`. `legacy/` is reference-only and
is excluded from active formatting/build/test checks.

## Protobuf

Regenerate protobuf code only when proto sources intentionally change:

```powershell
make proto
```

If `make` is unavailable, use the `protoc` command documented in the Makefile.

## Branch And Commit Style

Use focused branches and concise imperative commit messages, for example:

- `fix: handle gateway readiness timeout`
- `docs: update docker troubleshooting`
- `test: cover crc16 parser noise`

Run local CI equivalents before opening a PR:

```powershell
.\scripts\release-check.ps1
```
