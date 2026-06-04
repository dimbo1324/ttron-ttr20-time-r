# gRPC API

Step 5 adds a gRPC control plane for emulator and gateway services.

The gRPC API is a service-control/status API. It does not replace the FT1.2-like
TCP data path used by devices, emulator sessions, client polling, or gateway
polling.

## Files

Proto sources:

```text
proto/ft12/v1/common.proto
proto/ft12/v1/emulator.proto
proto/ft12/v1/gateway.proto
```

Generated Go package:

```text
github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1
```

Generation command:

```powershell
make proto
```

## EmulatorService

Default listen address: `:9100`.

RPCs:

- `GetStatus`
- `GetFaultMode`
- `SetFaultMode`
- `GetRecentEvents`

Run:

```powershell
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -mode sum -grpc-listen 127.0.0.1:9100
```

Disable gRPC:

```powershell
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -grpc-listen ""
```

## GatewayService

Default listen address: `:9200`.

RPCs:

- `GetStatus`
- `StartPolling`
- `StopPolling`
- `GetRecentEvents`
- `GetLastReadTime`

Run:

```powershell
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -mode sum -interval 1s -grpc-listen 127.0.0.1:9200
```

The gateway starts polling by default for backward compatibility. `StopPolling`
and `StartPolling` are idempotent control RPCs over the same internal polling
service.

## Notes

The HTTP API and Web UI are implemented as later layers on top of this gRPC
control plane. TLS, auth, database persistence, and OpenTelemetry tracing remain
future optional work.
