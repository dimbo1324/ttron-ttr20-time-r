# Roadmap

## The brief

> Написать программу на Go, которая каждую пятую секунду минуты вычитывает
> время прибора с датой и выводит его в лог — по FT1.2 через телепорт ТТР20.

`go run ./cmd/ft12-gateway`, with no flags, does exactly that: the shipped
default is an aligned schedule of one minute at an offset of five seconds. See
[Gateway](gateway.md) for the schedule modes and [Protocol](protocol.md) for
the frame format.

Everything below that line is scope this project took on deliberately, to make
the answer a bench someone can learn the protocol on rather than a script.

## Delivered

1. Go monorepo baseline.
2. FT1.2-like protocol core.
3. Emulator TCP service.
4. Gateway polling service.
5. gRPC contracts.
5.5. Architecture hardening.
6. HTTP API layer.
7. Docker, observability, and CI.
8. Final docs and release polish.
9. UI localisation, themes, infographics, exports, and documentation upgrade.
10. Gateway domain layer.
11. Web console.
12. Live control plane.

Step 5 is complete when proto contracts, generated Go code, emulator/gateway
gRPC adapters, client helpers, integration tests, docs, and build/test checks
are in place.

Step 5.5 is complete when command entrypoints are thin, app bootstrap packages
own process lifecycle, config loading is testable and validated, event IDs are
stable, gRPC mapper duplication is reduced, service files are decomposed, and
architecture checks are available.

Step 6 is complete when a thin Go HTTP API adapter exposes the gRPC control
plane as HTTP/JSON so clients can monitor status, inspect events, update
emulator fault mode, and start/stop gateway polling.

Step 7 is complete when Docker images, Docker Compose, health/readiness,
metrics, optional Prometheus scraping, CI quality gates, Makefile targets, and
operational docs are available without changing protocol or service business
logic.

Step 8 completes the MVP/portfolio scope when the README, docs index, release
flow, troubleshooting guide, examples, governance files, screenshots, doc-link
checks, release checks, and repository hygiene are ready for a public GitHub
presentation.

Step 9 is complete when the API provides JSON/CSV analysis exports and
documents the updated API behavior without changing the FT1.2-like wire
protocol.

Step 10 is complete when the gateway reasons about its own polling rather than
only performing it: wall-clock aligned scheduling, in-session retries separated
from reconnects, device clock skew with median classification and least-squares
drift, device health with hysteresis and latency percentiles, a device
inventory with a supervisor polling a fleet, and a command registry carrying
read-time and read-identity.

Step 11 is complete when the console can teach the protocol with nothing
installed: a browser-side FT1.2 implementation and bench engine, frame
analyzer, exchange monitor, dashboard, a reference written from first
principles, and a Russian/English interface throughout.

Step 12 is complete when the console can be pointed at a running Go stack and
read it — status, history windows, fleet — drive the emulator's fault mode, and
reconfigure a gateway that is already polling.

## Not built, on purpose

Two gaps are in the business logic rather than in the infrastructure around it,
and are the natural next pieces of work if the scope grows:

- **Per-device control.** The gRPC control plane binds to the primary device,
  so start, stop and settings act on that one device rather than on the fleet.
  `Supervisor.StartDevice` and `StopDevice` exist and are not exposed.
- **Outward alerting.** Clock and health transitions are logged and land in the
  event ring; nothing notifies anyone.

The rest is a change of scope rather than unfinished work:

- serial transport (RS-485) — the transport layer is TCP only;
- the wider FT1.2 command set — the registry carries two commands;
- persistence — no window, sample or event survives a restart;
- auth/RBAC, TLS/mTLS;
- Grafana dashboards, Kubernetes/Helm, OpenTelemetry tracing;
- advanced fuzzing;
- release binary workflow for Windows/Linux;
- production-grade secrets and config management.
