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

## Implemented Services

`cmd/ft12-emulator` runs a TCP emulator service. It accepts client/gateway
connections, parses FT1.2-like frames with the protocol core, handles read-time
requests, applies configured fault modes, and records in-memory status/history.

`cmd/ft12-gateway` runs a polling gateway service. It connects to an
emulator/device over TCP, periodically sends read-time requests, parses
responses, maintains status/history, and reconnects with backoff after errors.

Both emulator and gateway can also expose gRPC control APIs. The gRPC plane is
for service-to-service status/control and future API/Web layers; FT1.2-like TCP
remains the device data path.

`cmd/ft12-client` remains a simple direct polling/demo client.

`cmd/ft12-cli` is still a placeholder for future local inspection tools.

`cmd/ft12-api` runs a thin HTTP/JSON adapter. It talks to emulator and gateway
through the existing gRPC clients and exposes frontend-friendly DTOs for the Web
UI. It does not call emulator or gateway service packages directly.

The Web UI under `web/` is a React/Vite app that talks to `/api/v1` with
HTTP/JSON. gRPC remains an internal service API.

Docker Compose runs emulator, gateway, API, and web services on a private
Compose network. The web container serves the static Vite build with nginx and
proxies browser `/api` requests to the API service. The API exposes liveness,
readiness, and metrics endpoints for local operations and CI smoke tests.

## Step 5.5 Hardening

Command entrypoints are intentionally thin. Process bootstrap lives in:

- `internal/app/clientapp`;
- `internal/app/emulatorapp`;
- `internal/app/gatewayapp`;
- `internal/app/cliapp`.

Runtime orchestration for concurrent service runners uses
`internal/platform/lifecycle`. Logging remains compatible with `log.Logger`,
with `internal/platform/logging` as the structured logging migration point.

Recent frame/event history uses typed directions and service names, and the
event ring assigns stable monotonic IDs. This is important for future HTTP/Web
adapters that need stable event identity instead of slice-position IDs.

## Dependency Rules

```text
cmd/ft12-emulator -> internal/app/emulatorapp -> internal/emulator -> internal/transport/tcp -> internal/protocol
cmd/ft12-gateway  -> internal/app/gatewayapp  -> internal/gateway  -> internal/transport/tcp -> internal/protocol
cmd/ft12-client   -> internal/client   -> internal/protocol
app/* gRPC wiring -> internal/api/grpc  -> internal/{emulator,gateway}
cmd/ft12-api      -> internal/app/apiapp -> internal/api/http -> internal/api/grpc/client
web/              -> HTTP/JSON only
docker-compose    -> cmd services via flags; no protocol or business logic changes
```

`internal/protocol` depends only on the Go standard library. It does not depend
on TCP, emulator, gateway, logging, config, gRPC, HTTP, Web UI, or Docker.

`internal/observability/events` provides a small in-memory recent event ring for
service status and future UI/API layers. It is not a metrics stack.

See also:

- [Dependency rules](architecture/dependency-rules.md)
- [ADR 0005: Architecture hardening before Web UI](architecture/decisions/0005-architecture-hardening-before-web-ui.md)
- [HTTP API](http-api.md)
- [Web UI](web-ui.md)
- [Docker](docker.md)
- [Observability](observability.md)
- [CI](ci.md)
