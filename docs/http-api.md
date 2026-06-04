# HTTP API

Step 6 adds a thin HTTP/JSON adapter for browser and local tooling access.
Business logic remains in emulator and gateway services behind the existing
gRPC control plane.

## Run

```powershell
go run ./cmd/ft12-api -http-listen 127.0.0.1:8080 -emulator-grpc 127.0.0.1:9100 -gateway-grpc 127.0.0.1:9200
```

## Endpoints

- `GET /health`
- `GET /api/v1/health`
- `GET /api/v1/ready`
- `GET /metrics`
- `GET /api/v1/config`
- `GET /api/v1/overview`
- `GET /api/v1/emulator/status`
- `GET /api/v1/emulator/fault-mode`
- `PUT /api/v1/emulator/fault-mode`
- `GET /api/v1/emulator/events?limit=100`
- `GET /api/v1/gateway/status`
- `POST /api/v1/gateway/start`
- `POST /api/v1/gateway/stop`
- `GET /api/v1/gateway/last-read-time`
- `GET /api/v1/gateway/events?limit=100`
- `GET /api/v1/events?source=all&limit=100`
- `GET /api/v1/export/events.json?source=all&limit=100`
- `GET /api/v1/export/events.csv?source=all&limit=100`
- `GET /api/v1/export/overview.json?limit=50`
- `GET /api/v1/export/emulator-status.json`
- `GET /api/v1/export/gateway-status.json`

## Fault Mode Update

```json
{
  "responseDelayMs": 0,
  "corruptChecksum": false,
  "corruptChecksumProbability": 0,
  "fragmentResponse": false,
  "fragmentProbability": 0,
  "fragmentDelayMs": 40,
  "noResponse": false,
  "closeAfterRequest": false
}
```

## Error Model

Errors are returned consistently:

```json
{
  "error": {
    "code": "GATEWAY_UNAVAILABLE",
    "message": "gateway gRPC service is unavailable"
  }
}
```

The adapter maps malformed JSON and validation errors to `400`, unsupported
methods to `405`, upstream gRPC failures to `502`/`503`, and deadlines to `504`.

## Readiness

`GET /api/v1/ready` checks that the API process can reach both upstream gRPC
services. It returns `200` when emulator and gateway status calls succeed, and
`503` when either upstream is unavailable.

## Metrics

`GET /metrics` returns Prometheus-compatible text metrics for HTTP request
counts and total request duration by method, path, and status.

`GET /health` also includes build metadata fields: `version`, `commit`, and
`buildDate`.

## Exports

Export endpoints are read-only and use the same upstream gRPC status/events
data as the ordinary JSON API. They do not read local log files and do not
accept filesystem paths.

`source` supports `all`, `emulator`, or `gateway`. `limit` must be an integer
from `1` to `1000`; invalid values return `400 INVALID_LIMIT`.

Events CSV columns:

```text
timestamp,source,service,direction,command,checksumMode,remoteAddr,rawHex,message,error
```

CSV output is generated with Go's standard CSV writer, so commas, quotes, and
line breaks are escaped correctly. JSON downloads are indented and include an
`exportedAt` timestamp. Responses set `Content-Type` and `Content-Disposition`
headers with server-generated filenames such as
`ft12-events-YYYYMMDD-HHMMSS.csv`.

Exported files may contain protocol diagnostic data, raw hex, remote addresses,
and service counters. Treat them as local troubleshooting artifacts.

## Architecture

HTTP handlers depend on small interfaces implemented by gRPC client adapters.
They do not import emulator or gateway service packages directly, and they do
not duplicate protocol or polling behavior.
