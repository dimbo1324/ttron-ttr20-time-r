# Emulator Service

The emulator is a TCP service that accepts FT1.2-like frames and responds to the
read-time command.

Run:

```powershell
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -mode sum
```

CRC16:

```powershell
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -mode crc16
```

Fault examples:

```powershell
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -response-delay 2s
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -bad-checksum
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -fragment-response
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -no-response
```

The service keeps in-memory status counters and a fixed-size recent event
buffer. There is no database or external observability stack in this milestone.

The emulator also exposes a gRPC control API when `-grpc-listen` is set.
