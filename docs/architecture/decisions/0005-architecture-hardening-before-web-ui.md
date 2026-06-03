# 0005 Architecture Hardening Before Web UI

## Status

Accepted in Step 5.5.

## Context

The project already has a protocol core, TCP emulator, gateway poller, and gRPC
control plane. Before adding a Web UI, the runtime needs clearer bootstrap
boundaries, testable configuration loading, stable event IDs, and dependency
checks.

## Decision

- Keep `cmd/*` entrypoints thin and move process bootstrap to `internal/app/*`.
- Parse config through `LoadX(args []string)` with `Normalize` and `Validate`.
- Use a small lifecycle group for app-level concurrent runners.
- Keep generated protobuf code in `internal/api/grpc/ft12/v1`.
- Extract shared gRPC mapper code to `internal/api/grpc/mapping`.
- Add typed event directions/service names and ring-assigned monotonic IDs.
- Split emulator and gateway service files by facade, session/poller, history,
  and status-store responsibilities.
- Keep logging compatible with `log.Logger` while introducing
  `internal/platform/logging` as the migration point for future structured
  logging.
- Add architecture check scripts and a `make verify` target.

## Consequences

The Web UI can consume stable event IDs and can later sit behind an HTTP
adapter without pulling protocol or transport code into UI-facing packages.

Full package renaming from `internal/api/grpc/*` to `internal/adapters/grpc/*`
was deferred to avoid unnecessary churn while the generated protobuf package
still lives under `internal/api/grpc/ft12/v1`.

This step intentionally does not add Web UI, HTTP API, Docker, CI, persistence,
metrics, authentication, TLS, serial transport, or new protocol commands.
