# Development

## Go Version

The root module uses Go `1.26`.

The repository is intentionally a single root Go module. `go.work` is not used
for the active baseline.

## Commands

```powershell
go fmt ./...
.\scripts\check-go-format.ps1
go test ./...
go build ./...
.\scripts\check-architecture.ps1
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -mode sum
go run ./cmd/ft12-client -host 127.0.0.1 -port 9000 -crc sum
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -mode sum -interval 5s
go run ./cmd/ft12-api -http-listen 127.0.0.1:8080
make verify
make proto
docker compose config
docker compose up --build
```

Individual builds:

```powershell
go build -o bin/ft12-client ./cmd/ft12-client
go build -o bin/ft12-emulator ./cmd/ft12-emulator
go build -o bin/ft12-gateway ./cmd/ft12-gateway
go build -o bin/ft12-cli ./cmd/ft12-cli
go build -o bin/ft12-api ./cmd/ft12-api
```

The `Makefile` exposes the same common operations for environments with `make`.
`make verify` runs formatting, architecture checks, tests, and build.
`make clean-runtime-dry-run` previews ignored runtime/build cleanup, and
`make clean-runtime` removes those ignored local artifacts.

Useful service runs:

```powershell
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -mode crc16
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -mode crc16 -interval 1s
```

gRPC control:

```powershell
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -grpc-listen 127.0.0.1:9100
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -grpc-listen 127.0.0.1:9200
```

Web dashboard:

```powershell
cd web
npm install
npm run dev
```

Open `http://localhost:5173`. Vite proxies `/api` to the local HTTP API.

Docker Compose:

```powershell
docker compose up --build
docker compose down -v
```

Open `http://localhost:5173`. The nginx web container proxies `/api`, `/health`,
and `/metrics` to the API service.

## Active And Legacy Code

Active Go code lives in `cmd/` and `internal/`.

Legacy/reference code lives in:

- `legacy/python`;
- `legacy/go_sln`.

Legacy code is preserved for comparison and is not part of normal root module
build/test workflows.

Go formatting checks apply to active Go files and intentionally exclude
`legacy/`. The repository includes `.gitattributes` so active text files use LF
in Git. The dedicated Go format scripts compare `gofmt` output after CRLF/LF
normalization, which keeps Windows CI stable without weakening active-code
formatting.

## Local Logs And Cleanup

Runtime logs default to `runtime/logs`:

- `ft12-emulator.log`
- `ft12-gateway.log`
- `ft12-api.log`

The `-log` flag overrides the path; `-log=` sends logs to stdout. Do not log
secrets or request bodies. Protocol frame hex is logged for local diagnostics.

Cleanup scripts remove ignored runtime/build artifacts only:

```powershell
.\scripts\clean-runtime.ps1 -DryRun
.\scripts\clean-runtime.ps1
```

```sh
bash scripts/clean-runtime.sh --dry-run
bash scripts/clean-runtime.sh
```

## Architecture Checks

Dependency boundary scripts live in `scripts/`:

```powershell
.\scripts\check-architecture.ps1
```

```sh
sh scripts/check-architecture.sh
```

The checks keep `internal/protocol` independent from transports, config,
logging, service packages, gRPC adapters, and future adapter layers. They also
ensure active code does not import `legacy/`.

## Local CI Flow

When all tools are available, use:

```powershell
go fmt ./...
.\scripts\check-go-format.ps1
.\scripts\check-architecture.ps1
go test ./...
go build ./...
cd web
npm ci
npm run typecheck
npm run lint
npm run build
cd ..
docker compose config
docker compose build
.\scripts\check-doc-links.ps1
.\scripts\release-check.ps1
```
