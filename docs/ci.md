# CI

Step 7 adds GitHub Actions workflow `.github/workflows/ci.yml`.

## Triggers

- Pushes to repository branches, including `main`.
- Pull requests targeting `main`.

Workflow permissions are read-only:

```yaml
permissions:
  contents: read
```

Concurrency cancels older runs for the same Git ref.

## Jobs

| Job | Runner | Purpose |
| --- | --- | --- |
| `backend` | Ubuntu, Windows, macOS | Go formatting, tests, build |
| `architecture` | Ubuntu | dependency boundary script |
| `cleanup` | Ubuntu, Windows | cleanup script dry-run |
| `frontend` | Ubuntu | npm ci, typecheck, lint, build |
| `docker` | Ubuntu | compose config, compose build, compose smoke |
| `race` | Ubuntu | `go test -race ./...` |

## Local Equivalents

```powershell
go fmt ./...
.\scripts\check-go-format.ps1
.\scripts\check-architecture.ps1
go test ./...
go build ./...
cd web
npm ci
npm run typecheck
npm run lint
npm run build
cd ..
docker compose config
docker compose build
docker compose up -d
.\scripts\check-doc-links.ps1
.\scripts\clean-runtime.ps1 -DryRun
.\scripts\release-check.ps1
```

On Unix-like systems:

```sh
sh scripts/check-architecture.sh
sh scripts/check-go-format.sh
bash scripts/clean-runtime.sh --dry-run
```

## Formatting And Line Endings

CI uses `scripts/check-go-format.ps1` for the cross-platform backend matrix.
The script checks active Go files only and compares `gofmt` output after
normalizing CRLF/LF, so `windows-latest` does not fail merely because checkout
line endings differ. Real `gofmt` changes still fail the job.

`legacy/` is reference-only and excluded from active formatting/build/test
policy. Active code under `cmd/` and `internal/` is checked strictly.

`.gitattributes` pins LF for active text formats used by this repository,
including Go, proto, YAML, shell, PowerShell, Markdown, and TypeScript.

## Notes

The workflow intentionally avoids heavy lint suites until they are configured
with project-specific rules. Docker smoke is limited to stable health and HTTP
API checks to reduce flakiness. Release checks are documented for maintainers
and can be split into CI jobs later if needed.
