# TTRON TTR20 Time / FT1.2 Protocol Platform

A Go-based industrial protocol playground for FT1.2-like/TTR20 time-reading
communication. The current baseline provides a TCP emulator, a polling client,
shared protocol helpers, and a monorepo structure for future gateway, gRPC, Web
UI, Docker, CI, and observability work.

## Current Status

Implemented through Step 5.5:

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
- future CLI command placeholder;
- proto, web, deploy, docs, and legacy scaffolding;
- Python and old Go implementations preserved under `legacy/`.

Planned but not implemented yet:

- Web UI;
- Docker, CI, metrics, tracing, and release polish.

## Repository Layout

```text
cmd/                  active command entrypoints
  ft12-client/        polling client
  ft12-emulator/      TCP device emulator
  ft12-gateway/       future gateway placeholder
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
web/                  future Web UI
deploy/               future Docker/Compose assets
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
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -interval 5s
make proto
make verify
```

A root `Makefile` is also provided for Unix-like shells and environments with
`make`.

## Documentation

- [Architecture](docs/architecture.md)
- [Dependency Rules](docs/architecture/dependency-rules.md)
- [ADR 0005: Architecture Hardening Before Web UI](docs/architecture/decisions/0005-architecture-hardening-before-web-ui.md)
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
