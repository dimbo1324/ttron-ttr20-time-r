# Gateway Service

The gateway is a TCP polling service. It connects to an FT1.2-like
emulator/device, sends read-time requests, parses responses, tracks status, and
reconnects with backoff after errors.

Run with the emulator:

```powershell
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -mode sum
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -mode sum -interval 5s
```

CRC16:

```powershell
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -mode crc16
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -mode crc16 -interval 5s
```

Useful flags:

- `-target`: TCP address of the emulator/device;
- `-interval`: polling interval;
- `-timeout`: request/response timeout;
- `-connect-timeout`: TCP connect timeout;
- `-backoff-initial`, `-backoff-max`: reconnect behavior;
- `-recent`: in-memory recent event buffer size.

The gateway is local service logic only. gRPC APIs, Web UI, persistence, and
metrics are planned later.
