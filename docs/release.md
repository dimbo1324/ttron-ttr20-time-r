# Release

This project is currently prepared for an initial portfolio/MVP release. The
recommended first public version tag is `v0.1.0`.

## Versioning

Use semantic versioning for public tags:

- `v0.1.0`: initial portfolio release.
- Patch versions: documentation, CI, bug fixes that do not change protocol/API contracts.
- Minor versions: new optional features or non-breaking API additions.
- Major versions: breaking protocol, API, or deployment changes.

Go binaries expose build metadata through `/health` for the API. Docker builds
can pass `VERSION`, `COMMIT`, and `BUILD_DATE` build args.

## Local Release Check

PowerShell:

```powershell
.\scripts\release-check.ps1
```

Unix-like shell:

```sh
sh scripts/release-check.sh
```

The scripts run Go checks, architecture checks, frontend checks, Compose config,
and documentation link checks. Docker image build and smoke tests are still
recommended before tagging when Docker is available.

## Manual Docker Smoke

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

Optional Prometheus config:

```powershell
docker compose --profile observability config
```

## GitHub Release Flow

1. Ensure `main` is green in GitHub Actions.
2. Run the local release check.
3. Run Docker Compose smoke if Docker is available.
4. Review [CHANGELOG.md](../CHANGELOG.md).
5. Choose the version, for example `v0.1.0`.
6. Tag:

```powershell
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

7. Create a GitHub release from the tag.
8. Include release notes from [CHANGELOG.md](../CHANGELOG.md).
9. Attach built binaries only if a dedicated release-binary workflow exists.
10. Verify the Docker quick start from a fresh clone.

## Release Notes Checklist

- Scope summary.
- Supported commands/services.
- Docker Compose quick start.
- Known limitations.
- Security notes: no auth/TLS/persistence yet.
- Links to docs and examples.
