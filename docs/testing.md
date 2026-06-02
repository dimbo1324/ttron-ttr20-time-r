# Testing

Step 1 includes baseline tests for:

- additive checksum;
- CRC-16/Modbus;
- frame build, verify, extract, and payload helpers;
- hex dump formatting.

Required baseline checks:

```powershell
go fmt ./...
go test ./...
go build ./...
```

Future milestones should add protocol vectors, malformed-frame coverage,
emulator integration tests, gateway polling tests, and service contract tests.
