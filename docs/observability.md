# Observability

Step 7 adds a small observability baseline that is stable enough for local
Docker and CI smoke tests.

## Health

`GET /health` and `GET /api/v1/health` report that the API process is alive.
They do not require emulator or gateway upstreams.

Example:

```json
{
  "status": "ok",
  "service": "ft12-api",
  "version": "dev",
  "commit": "unknown",
  "buildDate": "unknown"
}
```

## Readiness

`GET /api/v1/ready` checks the API process and short upstream status requests to
the emulator and gateway gRPC services.

Ready response:

```json
{
  "status": "ready",
  "emulator": "ok",
  "gateway": "ok"
}
```

If either upstream is unavailable, the endpoint returns `503` with
`status: "not_ready"`.

## Metrics

`GET /metrics` exposes a minimal Prometheus-compatible text endpoint from
`ft12-api`.

Current metrics:

- `ft12_http_requests_total`
- `ft12_http_request_duration_seconds_total`

Both metrics are labeled by HTTP method, path, and status code.

## Prometheus

An optional Prometheus profile is available:

```powershell
docker compose --profile observability up -d --build
```

Open `http://localhost:9090`. The config lives at
`deploy/observability/prometheus.yml` and scrapes `ft12-api:8080/metrics`.

## Logs

Services emit key-value style logs for startup config, listen addresses,
request IDs, request method/path/status/duration, errors, and shutdown. Full
structured logging and tracing are intentionally deferred.

## Security Notes

Observability endpoints are unauthenticated in this local development baseline.
Do not expose the API or Web UI directly to untrusted networks.
