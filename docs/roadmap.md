# Roadmap

1. Go monorepo baseline. Implemented.
2. FT1.2 protocol core. Implemented.
3. Emulator TCP service. Implemented.
4. Gateway polling service. Implemented.
5. gRPC contracts. Implemented.
5.5. Architecture hardening before Web UI. Implemented.
6. Web UI and HTTP API layer. Implemented.
7. Docker, observability, and CI. Implemented.
8. Final docs and release polish. Implemented.
9. UI localisation, themes, infographics, exports, and documentation upgrade. Implemented.

Step 5 is complete when proto contracts, generated Go code, emulator/gateway
gRPC adapters, client helpers, integration tests, docs, and build/test checks
are in place.

Step 5.5 is complete when command entrypoints are thin, app bootstrap packages
own process lifecycle, config loading is testable and validated, event IDs are
stable, gRPC mapper duplication is reduced, service files are decomposed, and
architecture checks are available.

Step 6 is complete when a thin Go HTTP API adapter exposes the gRPC control
plane as HTTP/JSON, and a React/Vite Web UI can monitor status, inspect events,
update emulator fault mode, and start/stop gateway polling.

Step 7 is complete when Docker images, Docker Compose, health/readiness,
metrics, optional Prometheus scraping, CI quality gates, Makefile targets, and
operational docs are available without changing protocol or service business
logic.

Step 8 completes the current MVP/portfolio scope when the README, docs index,
release flow, troubleshooting guide, examples, governance files, screenshots,
doc-link checks, release checks, and repository hygiene are ready for a public
GitHub presentation.

Step 9 is complete when the Web UI is bilingual with Russian default, supports
persisted dark/light themes, uses a more compact responsive dashboard, includes
protocol and diagnostics infographics, provides JSON/CSV analysis exports, and
documents the updated UI/API behavior without changing the FT1.2-like wire
protocol.

## Completed MVP Scope

- Go monorepo baseline.
- FT1.2-like protocol core.
- TCP emulator service.
- Gateway polling service.
- gRPC control plane.
- Architecture hardening.
- HTTP API and Web UI.
- Docker, CI, observability.
- Final documentation, release, examples, and GitHub polish.
- Bilingual themed Web UI, protocol infographics, diagnostics, and local exports.

## Future Optional Milestones

These are intentionally outside the current MVP:

- auth/RBAC;
- TLS/mTLS;
- persistence;
- multi-device gateway management;
- serial transport;
- expanded FT1.2 command set;
- Grafana dashboard;
- Kubernetes/Helm;
- OpenTelemetry tracing;
- advanced fuzzing;
- Windows/Linux release binary workflow;
- production-grade secrets/config management.
