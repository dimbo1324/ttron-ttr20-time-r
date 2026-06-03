# Dependency Rules

Step 5.5 adds explicit architecture checks before Web UI work.

## Core Rule

`internal/protocol` is the wire-protocol core. It must stay independent from:

- network transports;
- config and logging packages;
- emulator and gateway service packages;
- gRPC adapters;
- future HTTP/Web adapters;
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

## Legacy Boundary

Active code under `cmd/`, `internal/`, and `proto/` must not import `legacy/`.
Legacy implementations are retained only as reference material.

## Local Check

Run:

```powershell
.\scripts\check-architecture.ps1
```

or:

```sh
sh scripts/check-architecture.sh
```

With `make`:

```sh
make check-architecture
make verify
```
