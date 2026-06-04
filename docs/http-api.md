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

## Architecture

HTTP handlers depend on small interfaces implemented by gRPC client adapters.
They do not import emulator or gateway service packages directly, and they do
not duplicate protocol or polling behavior.
