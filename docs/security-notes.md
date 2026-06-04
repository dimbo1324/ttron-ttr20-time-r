# Security Notes

This project is a local simulation and portfolio platform. It is not a hardened
production deployment.

## Current Scope

- No authentication or RBAC.
- No TLS or mTLS.
- No persistence or secrets management.
- No production network policy.
- No Kubernetes/Helm/cloud deployment.

## Existing Baseline

- Docker Go services use multi-stage builds and distroless non-root runtime
  images.
- The Web UI serves static files from nginx as a non-root user.
- Build contexts exclude `.env`, logs, binaries, `node_modules`, and build
  outputs.
- The HTTP API keeps recovery middleware and request IDs.
- The HTTP API sets `X-Content-Type-Options`, `X-Frame-Options`, and
  `Referrer-Policy` headers.
- JSON control endpoints use request body size limits.
- Export endpoints validate `limit`, use server-generated filenames, and do not
  accept local filesystem paths.
- Health/readiness/metrics endpoints are intended for local development and CI.
- Runtime logs default to ignored `runtime/logs` files. Logs should not contain
  secrets or request bodies; protocol frame hex is local diagnostic data.
  Exported JSON/CSV may also contain protocol diagnostic data and should stay
  local unless reviewed.

## Do Not Expose Publicly

Do not expose the API, Web UI, gRPC ports, or emulator TCP port to untrusted
networks without adding authentication, TLS, authorization, rate limiting, and
deployment-specific hardening.

## Future Hardening

Possible future work:

- Auth/RBAC for control endpoints.
- TLS/mTLS for API and gRPC.
- Secrets/config management.
- Read-only/user-scoped demo modes.
- Container image scanning.
- SBOM generation.
- Signed release artifacts.
