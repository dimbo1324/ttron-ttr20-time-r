# Changelog

## Unreleased

The gateway learned to reason about its own polling, the console learned to
read a real one, and the repository learned to check itself.

### Added

- **Gateway domain.** Wall-clock aligned scheduling, in-session retries kept
  separate from reconnects, device clock skew classified on a median with a
  least-squares drift rate and its fit, device health with hysteresis and
  latency percentiles, a device inventory with a supervisor polling a fleet,
  and a command registry carrying read-time and read-identity.
- **Web console** (`web/`), rebuilt on Next.js: a browser-side FT1.2
  implementation and bench engine that teaches the protocol with nothing
  installed, a frame analyzer, an exchange monitor, a dashboard, a reference
  written from first principles, and a Russian/English interface throughout.
- **Live mode.** The console can be pointed at a running Go stack: it reads
  status, the rolling history windows and the fleet, drives the emulator's
  fault mode, and reconfigures a gateway that is already polling.
- **Control plane.** `GetFleet`, `GetHistory` and `UpdateSettings` over gRPC,
  with `GET /api/v1/gateway/{fleet,history}` and `PUT /api/v1/gateway/settings`
  over HTTP. Gateway status now carries the schedule, retry budget, measured
  clock, device health and nameplate.
- **Docker for the whole project.** An image for the console, a development
  overlay with hot reload, per-Dockerfile build contexts, and base images
  pinned by digest.
- **Repository checks** (`tools/checks`): dependency boundaries read from the
  real import graph, gofmt, local Markdown links, artefact cleanup, and a
  release gate — one Go command replacing five pairs of shell and PowerShell
  scripts.
- **CI**: the frontend (lint, typecheck, tests, build), golangci-lint,
  ESLint, govulncheck, npm audit, CodeQL, Dependabot, and a release workflow
  that builds binaries for four platforms, publishes images with an SBOM, and
  cuts a GitHub release from this file.

### Changed

- The default schedule is the brief this project answers: with no flags, the
  gateway reads the device clock on the fifth second of every minute.
- Containers log to stdout because they are told to, not because writing to a
  file failed.
- The console proxies to the API through a route handler rather than a
  `rewrites()` entry, so `FT12_API_URL` is read per request and one image runs
  anywhere.
- Go 1.26.8, gRPC 1.83.2 and `golang.org/x/net` 0.58.0, which between them
  closed 24 reachable advisories.

### Fixed

- Device time was encoded in local time and decoded as UTC, so a healthy clock
  reported a skew equal to the host's timezone offset.
- An upstream timeout was answered `502 Bad Gateway` instead of `504`, because
  a wrapped `context.DeadlineExceeded` was compared with `==`.
- The health window discarded successful polls that completed in under a
  millisecond.
- Gateway configuration normalisation silently repaired invalid values instead
  of rejecting them.
- A gateway address was built with `fmt.Sprintf("%s:%d")`, which produces an
  unusable address for IPv6.

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
