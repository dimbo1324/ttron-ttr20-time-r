# Dependency Rules

The boundaries the architecture check enforces, and why each one is there.

## Core Rule

`internal/protocol` is the wire-protocol core. It must stay independent from:

- network transports;
- config and logging packages;
- emulator and gateway service packages;
- gRPC adapters;
- HTTP adapters;
- app bootstrap packages;
- HTTP adapters;
- deployment or observability stacks.

This keeps FT1.2-like frame parsing, checksums, commands, and codecs reusable
for future transports and tests.

## Service Boundaries

`internal/emulator` must not import `internal/gateway`.

`internal/gateway` must not import `internal/emulator`.

Both services can depend on shared protocol, transport, config, and
observability/event helpers.

## Adapter Boundaries

Generated protobuf code remains under `internal/api/grpc/ft12/v1`.

Handwritten gRPC adapters map service snapshots into protobuf DTOs through
`internal/api/grpc/mapping`. The shared mapper prevents duplicate checksum,
direction, event, timestamp, and service-state logic.

The HTTP API lives under `internal/api/http` and may import gRPC clients,
protobuf DTOs, config, and platform helpers. It must not import
`internal/emulator` or `internal/gateway` service packages directly.

## Legacy Boundary

Active code under `cmd/`, `internal/`, and `proto/` must not import `legacy/`.
Legacy implementations are retained only as reference material.

## Local Check

```sh
go run ./tools/checks architecture
```

It reads the real import graph from `go list`, so a forbidden package reached
through two intermediate hops fails as clearly as a direct import, and the
failure names the chain and the file and line the first hop is written on.
Production code is checked transitively; test files are checked for direct
imports only, because a test reaching a package through its own subject is not
a boundary violation.

`make check-architecture` and `make verify` run the same thing.
