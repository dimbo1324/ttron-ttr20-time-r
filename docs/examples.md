# Examples

More examples live under [examples/](../examples/).

## FT1.2-like Read-Time Frames

Read-time request in `sum` mode:

```text
68 03 68 00 01 01 02 16
```

Fields:

- `68`: start
- `03`: length of control/address/data
- `68`: repeated start
- `00`: request control
- `01`: adapter address
- `01`: read-time command
- `02`: additive checksum over `00 01 01`
- `16`: end

Read-time request in `crc16` mode:

```text
68 03 68 00 01 01 90 20 16
```

CRC bytes are little-endian CRC-16/Modbus over `00 01 01`.

Read-time response payload format:

```text
01 32 30 32 36 2d 30 36 2d 30 34 20 31 35 3a 30 34 3a 30 35
```

The first byte is command `0x01`; the rest is ASCII timestamp
`2026-06-04 15:04:05`.

## HTTP Examples

Health:

```powershell
curl http://127.0.0.1:8080/health
```

Readiness:

```powershell
curl http://127.0.0.1:8080/api/v1/ready
```

Overview:

```powershell
curl http://127.0.0.1:8080/api/v1/overview
```

Set normal fault mode:

```powershell
curl -X PUT http://127.0.0.1:8080/api/v1/emulator/fault-mode `
  -H "Content-Type: application/json" `
  -d "{\"responseDelayMs\":0,\"corruptChecksum\":false,\"corruptChecksumProbability\":0,\"fragmentResponse\":false,\"fragmentProbability\":0,\"fragmentDelayMs\":40,\"noResponse\":false,\"closeAfterRequest\":false}"
```

Stop and start gateway polling:

```powershell
curl -X POST http://127.0.0.1:8080/api/v1/gateway/stop
curl -X POST http://127.0.0.1:8080/api/v1/gateway/start
```

Metrics:

```powershell
curl http://127.0.0.1:8080/metrics
```

Events JSON export:

```powershell
curl "http://127.0.0.1:8080/api/v1/export/events.json?source=all&limit=100"
```

Events CSV export:

```powershell
curl "http://127.0.0.1:8080/api/v1/export/events.csv?source=all&limit=100"
```

Overview JSON export:

```powershell
curl http://127.0.0.1:8080/api/v1/export/overview.json
```

Gateway and emulator status exports:

```powershell
curl http://127.0.0.1:8080/api/v1/export/gateway-status.json
curl http://127.0.0.1:8080/api/v1/export/emulator-status.json
```
