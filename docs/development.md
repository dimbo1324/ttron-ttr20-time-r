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
```
