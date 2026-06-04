# Testing

The current baseline includes tests for:

- additive checksum;
- CRC-16/Modbus;
- additional CRC-16 vectors and checksum mode parsing edge cases;
- frame encode/decode, invalid length, truncated checksum, and typed validation errors;
- streaming parser fragmentation, noise, partial invalid-frame recovery, multiple-frame, invalid-frame, and max-size behavior;
- read-time command request/response parsing, including invalid timestamp length;
- codec read-time request/response helpers and wrong-command response handling;
- reusable TCP transport helpers;
- emulator integration behavior, fault modes, history, logging, and status counters;
- gateway polling, timeout behavior, backoff, history, and status counters;
- gRPC emulator and gateway control APIs;
- config loading and validation;
- lifecycle group cancellation/error behavior;
- stable event ring IDs and snapshot copy behavior;
- shared gRPC mapper behavior;
- HTTP API config, handlers, error shape, CORS, readiness, metrics, events merge, controls, security headers, JSON body limits, invalid JSON, and invalid limits;
- runtime logging helpers and cleanup dry-run scripts;
- hex dump formatting.

Required baseline checks:

```powershell
go fmt ./...
.\scripts\check-go-format.ps1
.\scripts\check-architecture.ps1
go test ./...
go build ./...
go build ./cmd/ft12-api
```

When available, also run:

```powershell
go test -race ./...
make verify
```

Frontend checks:

```powershell
cd web
npm ci
npm run typecheck
npm run build
npm run lint
```

Docker checks:

```powershell
docker compose config
docker compose build
docker compose up -d
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/api/v1/ready
curl http://127.0.0.1:8080/api/v1/overview
curl http://127.0.0.1:8080/api/v1/events
curl http://127.0.0.1:8080/metrics
docker compose down -v
```

Documentation and release checks:

```powershell
.\scripts\check-doc-links.ps1
.\scripts\clean-runtime.ps1 -DryRun
.\scripts\release-check.ps1
```

`make test-fuzz` documents the current fuzz entrypoint status. Fuzzing is not
mandatory yet because no stable fuzz corpus is configured for this milestone.

Manual smoke coverage should include client/emulator in `sum` and `crc16`,
gateway/emulator in `sum` and `crc16`, HTTP API health/readiness/status/events
endpoints, Docker Compose, and the Web UI dashboard in a browser when the
environment supports it.

Future optional milestones can add deeper Web UI/API contract coverage and
release artifact automation.
