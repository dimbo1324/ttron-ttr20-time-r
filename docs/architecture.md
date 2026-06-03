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

## Dependency Rules

```text
cmd/ft12-emulator -> internal/emulator -> internal/transport/tcp -> internal/protocol
cmd/ft12-gateway  -> internal/gateway  -> internal/transport/tcp -> internal/protocol
cmd/ft12-client   -> internal/client   -> internal/protocol
cmd/* gRPC wiring -> internal/api/grpc  -> internal/{emulator,gateway}
```

`internal/protocol` depends only on the Go standard library. It does not depend
on TCP, emulator, gateway, logging, config, gRPC, HTTP, Web UI, or Docker.

`internal/observability/events` provides a small in-memory recent event ring for
service status and future UI/API layers. It is not a metrics stack.
