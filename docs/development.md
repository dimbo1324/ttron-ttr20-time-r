# Development

## Go Version

The root module uses Go `1.26`.

The repository is intentionally a single root Go module. `go.work` is not used
for the active baseline.

## Commands

```powershell
go fmt ./...
go test ./...
go build ./...
go run ./cmd/ft12-emulator
go run ./cmd/ft12-client
```

Individual builds:

```powershell
go build -o bin/ft12-client ./cmd/ft12-client
go build -o bin/ft12-emulator ./cmd/ft12-emulator
go build -o bin/ft12-gateway ./cmd/ft12-gateway
go build -o bin/ft12-cli ./cmd/ft12-cli
```

The `Makefile` exposes the same common operations for environments with `make`.

## Active And Legacy Code

Active Go code lives in `cmd/` and `internal/`.

Legacy/reference code lives in:

- `legacy/python`;
- `legacy/go_sln`.

Legacy code is preserved for comparison and is not part of normal root module
build/test workflows.
