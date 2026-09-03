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

- Every image is a multi-stage build with a distroless Debian runtime running
  as `nonroot:nonroot`, the console included. A runtime image carries the
  binary and the health check and no toolchain.
- Base images are pinned by digest, not by tag.
- Each Dockerfile has its own ignore file, and none of them include `.env`,
  logs, binaries or build output.
- Every push and every Monday, `govulncheck` looks for Go advisories that are
  actually reachable from this code, `pnpm audit` looks at the console's
  production tree, and CodeQL runs over Go and TypeScript. See [CI](ci.md).
- A release publishes an SBOM and a provenance attestation with each image and
  scans them with Trivy into the repository's security tab.
- Dependabot opens updates for Go modules, npm, GitHub Actions and Docker base
  images.
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

Do not expose the API, gRPC ports, or emulator TCP port to untrusted
networks without adding authentication, TLS, authorization, rate limiting, and
deployment-specific hardening.

## Future Hardening

- Auth/RBAC for control endpoints.
- TLS/mTLS for the API and for gRPC.
- Secrets and config management.
- A read-only or user-scoped demo mode.
- Signed release artefacts. Releases carry checksums and a provenance
  attestation today; the binaries themselves are not signed.
