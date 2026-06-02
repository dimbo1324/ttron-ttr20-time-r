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
- hex dump formatting.

Required baseline checks:

```powershell
go fmt ./...
go test ./...
go build ./...
```

Manual smoke coverage should include client/emulator in `sum` and `crc16`, plus
gateway/emulator in `sum` and `crc16`.

Future milestones should add gRPC contract tests and broader service API tests.
