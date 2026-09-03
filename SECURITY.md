# Security Policy

## Supported Versions

The supported development version is the latest `main` branch. Tagged releases
may be supported once formal releases begin.

## Reporting Security Issues

This repository currently has no private security contact configured. For
non-sensitive issues, open a GitHub issue with a clear description and
reproduction steps.

If you believe a report is sensitive, avoid posting exploit details publicly.
Use GitHub's private vulnerability reporting feature if it is enabled for the
repository, or contact the repository owner through GitHub.

## Current Security Scope

This project is a local simulation and portfolio platform. It does not currently
include:

- authentication;
- RBAC;
- TLS/mTLS;
- persistence;
- production secrets management;
- production deployment hardening.

Do not expose the API, gRPC ports, or emulator TCP port to untrusted
networks without additional hardening.

## What Is Scanned

Every push, and once a week on a schedule, the repository runs `govulncheck`
over the Go module, `pnpm audit` over the console's production dependencies,
and CodeQL over both languages. Released images carry an SBOM and a provenance
attestation and are scanned with Trivy. Dependabot opens dependency updates
for Go, npm, GitHub Actions and Docker base images.

See [docs/ci.md](docs/ci.md) and [docs/security-notes.md](docs/security-notes.md).
