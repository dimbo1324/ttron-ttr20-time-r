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
- HTTP API config, handlers, error shape, CORS, readiness, metrics, events merge, and controls;
- hex dump formatting.

Required baseline checks:

```powershell
go fmt ./...
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

`make test-fuzz` documents the current fuzz entrypoint status. Fuzzing is not
mandatory yet because no stable fuzz corpus is configured for this milestone.

Manual smoke coverage should include client/emulator in `sum` and `crc16`,
gateway/emulator in `sum` and `crc16`, HTTP API health/readiness/status/events
endpoints, Docker Compose, and the Web UI dashboard in a browser when the
environment supports it.

Future milestones should add deeper Web UI/API contract coverage and release
artifact validation.
