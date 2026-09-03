# Documentation Index

## Start here

- [A tour of the console](tour.md) — what the project looks like, in pictures
- [Architecture](architecture.md) — the services and the boundaries between them
- [Roadmap](roadmap.md) — the brief, what was delivered, what was left undone
  on purpose

## The protocol and the services

- [Protocol](protocol.md) — frame format, checksums, commands, the time-zone trap
- [Emulator](emulator.md) — the device and its fault modes
- [Gateway](gateway.md) — schedules, retries, clock, health, inventories
- [gRPC API](grpc-api.md) — the internal control plane
- [HTTP API](http-api.md) — endpoints, status shapes, settings, exports
- [Web console](console.md) — the two data sources and what each can change

## Building and running

- [Development](development.md) — commands, checks, coverage, logs
- [Testing](testing.md) — what is covered and how to run it
- [Docker](docker.md) — the Compose stack, the console's proxy, the images
- [CI](ci.md) — what runs on every push and on every tag
- [Observability](observability.md) — health, readiness, metrics
- [Troubleshooting](troubleshooting.md) — the failures that actually happen
- [Release](release.md) — cutting a version
- [Security notes](security-notes.md) — what is and is not hardened
- [Repository checklist](repository-checklist.md) — before a demo or a release

## Reference

- [Examples](examples.md) — frames on the wire, and calls against the API
- [Dependency rules](architecture/dependency-rules.md) — the boundaries the
  architecture check enforces
- [ADR 0005: Architecture hardening before Web UI](architecture/decisions/0005-architecture-hardening-before-web-ui.md)
- [Legacy implementations](legacy.md) — kept for comparison, not for building
- [Documentation media](../tools/docmedia/README.md) — how the screenshots are
  produced

The original protocol PDF and the assignment materials are preserved under
`docs/files/` and `task/`.
