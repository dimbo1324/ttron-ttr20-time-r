# Changelog

## Unreleased

### Removed

- Frontend Web UI (`web/`), its Docker Compose service, CI job, Makefile
  targets, and related documentation. The platform is now backend-only and is
  driven through the HTTP/JSON API.

## v0.1.0 - Initial portfolio release

Planned initial release scope:

- Go monorepo baseline.
- FT1.2-like frame encoder/decoder.
- Additive checksum and CRC-16/Modbus support.
- Stream parser for fragmented TCP frames.
- Read-time command model.
- TCP emulator with fault modes and recent events.
- Gateway polling service with reconnect/backoff.
- gRPC control plane.
- HTTP/JSON API.
- React/Vite/Tailwind Web UI.
- Docker Compose stack with nginx proxy.
- Health, readiness, and metrics endpoints.
- Optional Prometheus profile.
- GitHub Actions CI.
- Release, troubleshooting, examples, and repository governance docs.
