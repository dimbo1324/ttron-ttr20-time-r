# Testing

The current baseline includes tests for:

- additive checksum;
- CRC-16/Modbus;
- checksum mode parsing and verification;
- frame encode/decode and typed validation errors;
- streaming parser fragmentation, noise, multiple-frame, invalid-frame, and max-size behavior;
- read-time command request/response parsing;
- codec read-time request/response helpers;
- reusable TCP transport helpers;
- emulator integration behavior and fault modes;
- gateway polling and timeout behavior;
- gRPC emulator and gateway control APIs;
- config loading and validation;
- lifecycle group cancellation/error behavior;
- stable event ring IDs and snapshot copy behavior;
- shared gRPC mapper behavior;
- hex dump formatting.

Required baseline checks:

```powershell
go fmt ./...
.\scripts\check-architecture.ps1
go test ./...
go build ./...
```

When available, also run:

```powershell
go test -race ./...
make verify
```

`make test-fuzz` documents the current fuzz entrypoint status. Fuzzing is not
mandatory yet because no stable fuzz corpus is configured for this milestone.

Manual smoke coverage should include client/emulator in `sum` and `crc16`, plus
gateway/emulator in `sum` and `crc16`, plus gRPC GetStatus/GetLastReadTime style
checks.

Future milestones should add Web UI/API contract coverage.
