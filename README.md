# TTRON TTR20 Time / FT1.2 Protocol Platform

A Go-based industrial protocol playground for FT1.2-like/TTR20 time-reading
communication. The current baseline provides a TCP emulator, a polling client,
shared protocol helpers, and a monorepo structure for future gateway, gRPC, Web
UI, Docker, CI, and observability work.

## Current Status

Implemented in Step 1:

- root Go monorepo module;
- active Go client command;
- active Go TCP emulator command;
- shared checksum, frame, logging, and utility packages;
- future gateway and CLI command placeholders;
- proto, web, deploy, docs, and legacy scaffolding;
- Python and old Go implementations preserved under `legacy/`.

Planned but not implemented yet:

- deeper FT1.2 protocol core;
- hardened emulator service;
- gateway polling service;
- protobuf/gRPC contracts;
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
  protocol/           checksum and frame helpers
  client/             polling client runtime
  emulator/           TCP emulator runtime
  config/             standard-library flag config
  logging/            baseline logger
  util/               shared helpers
proto/                future protobuf/gRPC contracts
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
```

Run the emulator:

```powershell
go run ./cmd/ft12-emulator -host 127.0.0.1 -port 9000 -crc sum -delay 100 -badcrc 0.2 -fragment 0.1 -adapter 1
```

Run the client in another terminal:

```powershell
go run ./cmd/ft12-client -host 127.0.0.1 -port 9000 -crc sum -adapter 1 -timeout 1200 -retries 2 -pollstep 1
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

- `-host`, `-port`: listen address;
- `-crc`: `sum` or `crc16`;
- `-delay`: fixed response delay in milliseconds;
- `-badcrc`: probability `[0..1]` of sending a corrupted checksum;
- `-fragment`: probability `[0..1]` of sending a fragmented response;
- `-adapter`: adapter address byte;
- `-readtimeout`: connection read timeout in seconds;
- `-log`: log file path, or stdout when empty.

## Development Commands

```powershell
go fmt ./...
go test ./...
go build ./...
go build -o bin/ft12-client ./cmd/ft12-client
go build -o bin/ft12-emulator ./cmd/ft12-emulator
go build -o bin/ft12-gateway ./cmd/ft12-gateway
go build -o bin/ft12-cli ./cmd/ft12-cli
```

A root `Makefile` is also provided for Unix-like shells and environments with
`make`.

## Documentation

- [Architecture](docs/architecture.md)
- [Protocol](docs/protocol.md)
- [Development](docs/development.md)
- [Testing](docs/testing.md)
- [Roadmap](docs/roadmap.md)
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
