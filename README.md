# TTRON TTR20 Time / FT1.2 Protocol Platform

A Go-based industrial protocol playground for FT1.2-like/TTR20 time-reading
communication. The current baseline provides a TCP emulator, a polling client,
shared protocol helpers, a Web UI, Docker Compose, CI, and observability
baseline.

## Current Status

Implemented through Step 7:

- root Go monorepo module;
- active Go client command;
- active Go TCP emulator command;
- typed FT1.2-like protocol core;
- checksum mode abstraction with `sum` and `crc16`;
- frame encoder, decoder, typed errors, and streaming parser;
- read-time command model and high-level codec helpers;
- industrial-style TCP emulator service with status, fault modes, sessions, and recent events;
- gateway polling service with TCP reconnect, retry/backoff, status, and recent events;
- reusable TCP transport helpers;
- protobuf/gRPC contracts and generated Go code;
- gRPC control APIs for emulator and gateway;
- thin command entrypoints with app bootstrap packages;
- config loading via testable `args []string` APIs with validation;
- stable typed event history IDs for future UI/API consumers;
- shared gRPC mapping helpers;
- architecture boundary check scripts;
- HTTP/JSON API adapter over internal gRPC clients;
- React/Vite/TypeScript Web UI dashboard;
- Docker and Docker Compose local stack;
- non-root container runtime images and healthchecks;
- API readiness and metrics endpoints;
- optional Prometheus scrape profile;
- GitHub Actions CI for backend, frontend, architecture, Docker, and race checks;
- future CLI command placeholder;
- proto, web, deploy, docs, and legacy scaffolding;
- Python and old Go implementations preserved under `legacy/`.

Planned but not implemented yet:

- final docs and release polish;
- auth, TLS, persistence, multi-device fleet features, Kubernetes/Helm, and cloud deployment.

## Repository Layout

```text
cmd/                  active command entrypoints
  ft12-client/        polling client
  ft12-emulator/      TCP device emulator
  ft12-gateway/       future gateway placeholder
  ft12-api/           HTTP/JSON API adapter
  ft12-cli/           future CLI placeholder
internal/             active Go packages
  app/                process bootstrap for command entrypoints
  protocol/           checksum, frame, command, and codec core
  client/             polling client runtime
  emulator/           TCP emulator service
  gateway/            polling gateway service
  api/grpc/           generated gRPC code and handwritten adapters
  platform/           lifecycle and logging platform helpers
  transport/          reusable TCP helpers
  config/             standard-library flag config
  logging/            baseline logger
  util/               shared helpers
proto/                protobuf/gRPC contract sources
web/                  React/Vite dashboard and nginx runtime image
deploy/               Docker and observability assets
docs/                 architecture, protocol, development, testing, roadmap
legacy/               retained reference implementations
task/                 original assignment document
```

## Quick Start

Build and test from the repository root:

```powershell
go build ./...
go test ./...
.\scripts\check-architecture.ps1
```

Run the emulator:

```powershell
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -mode sum
```

Run the client in another terminal:

```powershell
go run ./cmd/ft12-client -host 127.0.0.1 -port 9000 -crc sum -adapter 1 -timeout 1200 -retries 2 -pollstep 1
```

Run the gateway poller instead of the demo client:

```powershell
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -mode sum -interval 5s
```

Run with gRPC control APIs:

```powershell
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -mode sum -grpc-listen 127.0.0.1:9100
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -mode sum -interval 1s -grpc-listen 127.0.0.1:9200
```

Run the HTTP API:

```powershell
go run ./cmd/ft12-api -http-listen 127.0.0.1:8080 -emulator-grpc 127.0.0.1:9100 -gateway-grpc 127.0.0.1:9200
```

Run the Web UI:

```powershell
cd web
npm install
npm run dev
```

Open `http://localhost:5173`.

Run the full Docker Compose stack:

```powershell
docker compose up --build
```

Open:

- Web UI: `http://localhost:5173`
- API health: `http://localhost:8080/health`
- API readiness: `http://localhost:8080/api/v1/ready`
- API metrics: `http://localhost:8080/metrics`

Stop the stack:

```powershell
docker compose down -v
```

CRC16 mode:

```powershell
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -mode crc16
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -mode crc16 -interval 5s
```

Fault examples:

```powershell
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -bad-checksum
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -fragment-response
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -response-delay 2s
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -no-response
```

Log files can be enabled with `-log client.log` or `-log emulator.log`.

## Flags

Client:

- `-host`, `-port`: emulator/device address;
- `-crc`: `sum` or `crc16`;
- `-adapter`: adapter address byte;
- `-timeout`: response timeout in milliseconds;
- `-retries`: retry count;
- `-pollstep`: polling ticker step in seconds;
- `-log`: log file path, or stdout when empty.

Emulator:

- `-listen`: listen address, for example `127.0.0.1:9000`;
- `-host`, `-port`: legacy listen address flags;
- `-mode` / `-crc`: `sum` or `crc16`;
- `-delay`: fixed response delay in milliseconds;
- `-response-delay`: duration response delay;
- `-badcrc`: probability `[0..1]` of sending a corrupted checksum;
- `-bad-checksum`: always corrupt response checksum;
- `-fragment`: probability `[0..1]` of sending a fragmented response;
- `-fragment-response`: always fragment responses;
- `-no-response`: receive valid request but send no response;
- `-adapter`: adapter address byte;
- `-readtimeout`: connection read timeout in seconds;
- `-log`: log file path, or stdout when empty.
- `-grpc-listen`: gRPC control listen address, default `:9100`; empty disables gRPC.

Gateway:

- `-target`: emulator/device TCP address;
- `-mode` / `-crc`: `sum` or `crc16`;
- `-interval`: polling interval;
- `-timeout`: request/response timeout;
- `-connect-timeout`: TCP connect timeout;
- `-backoff-initial`, `-backoff-max`: reconnect backoff settings;
- `-recent`: recent event buffer size;
- `-log`: log file path, or stdout when empty.
- `-grpc-listen`: gRPC control listen address, default `:9200`; empty disables gRPC.

HTTP API:

- `-http-listen`: HTTP listen address, default `:8080`;
- `-emulator-grpc`: emulator gRPC address, default `127.0.0.1:9100`;
- `-gateway-grpc`: gateway gRPC address, default `127.0.0.1:9200`;
- `-timeout`: upstream request timeout, default `3s`;
- `-cors-origin`: allowed CORS origin, default `http://localhost:5173`;
- `-log`: log file path, or stdout when empty.

## HTTP API

Core endpoints:

- `GET /health`
- `GET /api/v1/ready`
- `GET /metrics`
- `GET /api/v1/overview`
- `GET /api/v1/emulator/status`
- `GET /api/v1/emulator/fault-mode`
- `PUT /api/v1/emulator/fault-mode`
- `GET /api/v1/gateway/status`
- `POST /api/v1/gateway/start`
- `POST /api/v1/gateway/stop`
- `GET /api/v1/events?source=all&limit=100`

See [HTTP API](docs/http-api.md), [Docker](docs/docker.md), and
[Observability](docs/observability.md).

## Docker Compose Services

| Service | Purpose | Host ports |
| --- | --- | --- |
| `ft12-emulator` | FT1.2-like TCP emulator and gRPC control | `9000`, `9100` |
| `ft12-gateway` | polling gateway and gRPC control | `9200` |
| `ft12-api` | HTTP/JSON adapter, readiness, metrics | `8080` |
| `ft12-web` | nginx static Web UI and `/api` proxy | `5173` |
| `prometheus` | optional metrics scrape profile | `9090` |

The Web UI uses relative `/api` requests. In Docker Compose, nginx proxies them
to `ft12-api:8080`, so the browser never needs to resolve Compose service DNS.

## Observability

- `GET /health`: API liveness.
- `GET /api/v1/ready`: API readiness plus emulator/gateway upstream checks.
- `GET /metrics`: minimal Prometheus-compatible HTTP request metrics.
- `docker compose --profile observability up -d --build`: optional Prometheus
  on `http://localhost:9090`.

Logs use key-value style fields for startup config, listen addresses, request
IDs, HTTP method/path/status/duration, errors, and shutdown.

## Development Commands

```powershell
go fmt ./...
go test ./...
go build ./...
.\scripts\check-architecture.ps1
go build -o bin/ft12-client ./cmd/ft12-client
go build -o bin/ft12-emulator ./cmd/ft12-emulator
go build -o bin/ft12-gateway ./cmd/ft12-gateway
go build -o bin/ft12-cli ./cmd/ft12-cli
go build -o bin/ft12-api ./cmd/ft12-api
go build -o bin/ft12-healthcheck ./cmd/ft12-healthcheck
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -interval 5s
make proto
make verify
docker compose config
docker compose build
```

Security note: the local API and Web UI do not implement auth or TLS yet. Do not
expose them to untrusted networks.

A root `Makefile` is also provided for Unix-like shells and environments with
`make`.

## Documentation

- [Architecture](docs/architecture.md)
- [Dependency Rules](docs/architecture/dependency-rules.md)
- [ADR 0005: Architecture Hardening Before Web UI](docs/architecture/decisions/0005-architecture-hardening-before-web-ui.md)
- [HTTP API](docs/http-api.md)
- [Web UI](docs/web-ui.md)
- [Docker](docs/docker.md)
- [Observability](docs/observability.md)
- [CI](docs/ci.md)
- [Protocol](docs/protocol.md)
- [Development](docs/development.md)
- [Testing](docs/testing.md)
- [Roadmap](docs/roadmap.md)
- [Emulator](docs/emulator.md)
- [Gateway](docs/gateway.md)
- [gRPC API](docs/grpc-api.md)
- [Legacy](docs/legacy.md)

The original PDF and task document are preserved under `docs/files/` and
`task/`.

## Legacy Code

`legacy/python` contains the Python prototype/reference implementation.
`legacy/go_sln` contains the historical separate Go client/server modules.
Neither legacy area participates in the root Go build/test flow.

## Safety Note

This project is a simulation and learning platform. It is not intended for
direct control of real industrial devices without independent validation,
hardware-specific review, and operational safety controls.

## License

MIT. See [LICENSE](LICENSE).
