# Architecture

## Target Platform

The long-term direction is an industrial FT1.2-like/TTR20 protocol gateway and
emulator platform:

- Go backend services;
- clean protocol core;
- TCP device emulator;
- polling gateway;
- gRPC service-to-service contracts;
- small Web UI;
- Docker, CI, tests, and operational documentation.

## Step 1 Baseline

Step 1 establishes the repository shape and keeps the current working behavior:

- `cmd/ft12-client` runs the existing polling client behavior;
- `cmd/ft12-emulator` runs the existing TCP emulator behavior;
- `internal/protocol` contains shared checksum and frame helpers;
- `internal/client` and `internal/emulator` contain active runtime logic;
- `cmd/ft12-gateway` and `cmd/ft12-cli` compile as placeholders only.

## Dependency Rules

The baseline intentionally uses the Go standard library only. Heavy config,
logging, gRPC, web, database, metrics, and container dependencies are deferred
until their dedicated milestones.

Active source belongs under `cmd/` and `internal/`. Reference source belongs
under `legacy/`.
