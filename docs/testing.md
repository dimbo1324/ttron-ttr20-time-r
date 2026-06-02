# Testing

The current baseline includes tests for:

- additive checksum;
- CRC-16/Modbus;
- checksum mode parsing and verification;
- frame encode/decode and typed validation errors;
- streaming parser fragmentation, noise, multiple-frame, invalid-frame, and max-size behavior;
- read-time command request/response parsing;
- codec read-time request/response helpers;
- hex dump formatting.

Required baseline checks:

```powershell
go fmt ./...
go test ./...
go build ./...
```

Future milestones should add protocol vectors, malformed-frame coverage,
emulator integration tests, gateway polling tests, and service contract tests.
