# curl Examples

Start the stack first:

```powershell
docker compose up -d --build
```

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

Events:

```powershell
curl "http://127.0.0.1:8080/api/v1/events?source=all&limit=20"
```

Emulator status:

```powershell
curl http://127.0.0.1:8080/api/v1/emulator/status
```

Gateway status:

```powershell
curl http://127.0.0.1:8080/api/v1/gateway/status
```

Set fault mode to normal:

```powershell
curl -X PUT http://127.0.0.1:8080/api/v1/emulator/fault-mode `
  -H "Content-Type: application/json" `
  -d "{\"responseDelayMs\":0,\"corruptChecksum\":false,\"corruptChecksumProbability\":0,\"fragmentResponse\":false,\"fragmentProbability\":0,\"fragmentDelayMs\":40,\"noResponse\":false,\"closeAfterRequest\":false}"
```

Stop/start gateway polling:

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

Service status exports:

```powershell
curl http://127.0.0.1:8080/api/v1/export/gateway-status.json
curl http://127.0.0.1:8080/api/v1/export/emulator-status.json
```
