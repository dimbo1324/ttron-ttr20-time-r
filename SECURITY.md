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

Do not expose the API, Web UI, gRPC ports, or emulator TCP port to untrusted
networks without additional hardening.
