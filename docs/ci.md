# CI

Three workflows. `ci.yml` runs on every push and every pull request against
`main`; `security.yml` runs on those and once a week besides; `release.yml`
runs on a tag.

Workflow permissions are read-only except where a job needs more and says so
— the release workflow to publish, CodeQL to upload its results. Concurrency
cancels older runs for the same ref.

## Every push

| Job | Runner | What it establishes |
| --- | --- | --- |
| `backend` | Ubuntu, Windows, macOS | `gofmt`, `go test ./...`, `go build ./...` |
| `frontend` | Ubuntu | ESLint, `tsc --noEmit`, Jest, the production build |
| `coverage` | Ubuntu | both suites measured, a floor enforced, the reports kept |
| `lint` | Ubuntu | golangci-lint |
| `architecture` | Ubuntu | dependency boundaries, and that no source is hidden from git |
| `docs` | Ubuntu | every local Markdown link resolves |
| `cleanup` | Ubuntu, Windows | the cleanup command's dry-run does not misfire |
| `docker` | Ubuntu | `compose config`, `compose build`, and a smoke test of the running stack |
| `race` | Ubuntu | `go test -race ./...` |

The backend matrix is three operating systems because this project is
developed on Windows and deployed on Linux, and the differences that bite —
line endings, path separators, `make` not existing — bite at the boundary
between them.

The `docker` job brings the stack up and asks it questions: health, readiness,
the overview, the events feed, and the gateway fleet **through the console's
own proxy** on port 3000. That last one is the check that the console's
`/upstream` handler resolves `FT12_API_URL` at run time; a build-time rewrite
would pass every other test in this list and fail in the image.

The `coverage` job publishes the profile, the HTML view and the console's lcov
as an artifact, and prints the summary into the job's own page. The floor is a
floor rather than a ratchet: a number that must never fall turns every
refactor into a negotiation with the metric, whereas a floor says only that
the suite has not quietly stopped covering the code.

## Every push, and Mondays

| Job | What it looks for |
| --- | --- |
| `govulncheck` | Go advisories that are actually reachable from this code |
| `npm-audit` | production advisories in the console's dependency tree |
| `codeql` | Go and TypeScript, on the default query set |

The weekly schedule is the point of the workflow. Code that has not changed
still becomes vulnerable; the advisory arrives on a Monday whether or not
anyone pushed.

Dependabot (`.github/dependabot.yml`) opens the updates: Go modules, npm
grouped so that a Next.js major arrives on its own rather than inside a batch,
GitHub Actions, and Docker base images.

## Every tag

`release.yml`, on `v*`:

1. **verify** — the full gate again. A tag is a promise, and CI running
   afterwards is a promise made before it was checked.
2. **binaries** — linux/amd64, linux/arm64, windows/amd64, darwin/arm64, with
   SHA-256 sums.
3. **images** — multi-architecture images to GHCR with an SBOM and a
   provenance attestation, scanned with Trivy into the security tab.
4. **release** — the GitHub release, with the notes taken from this version's
   section of [CHANGELOG.md](../CHANGELOG.md).

It can be rehearsed without spending a version number: run it by hand from the
Actions tab with **Publish** left off, and it builds and scans everything and
pushes nothing. See [Release](release.md).

## The same thing locally

The checks are one Go program, so the local command is the CI command:

```sh
go run ./tools/checks release
```

Or a stage at a time:

```sh
go fmt ./...
go run ./tools/checks format
go run ./tools/checks architecture
go run ./tools/checks doc-links
go vet ./...
go test ./...
go build ./...
make lint
pnpm --dir web lint && pnpm --dir web typecheck && pnpm --dir web test
docker compose config
```

## Formatting and line endings

The format check runs across the whole backend matrix and compares `gofmt`
output after normalising CRLF/LF, so `windows-latest` does not fail merely
because the checkout used different line endings. Real formatting drift still
fails.

`legacy/` is reference material and is excluded from active formatting, build
and test policy. `.gitattributes` pins LF for Go, proto, YAML, shell,
PowerShell, Markdown and TypeScript.
