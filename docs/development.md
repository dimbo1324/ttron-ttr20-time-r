# Development

## Go Version

The root module uses Go `1.26`.

The repository is intentionally a single root Go module. `go.work` is not used
for the active baseline.

## Commands

```powershell
go fmt ./...
.\scripts\check-go-format.ps1
go test ./...
go build ./...
.\scripts\check-architecture.ps1
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -mode sum
go run ./cmd/ft12-client -host 127.0.0.1 -port 9000 -crc sum
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -mode sum -schedule interval -interval 5s
go run ./cmd/ft12-api -http-listen 127.0.0.1:8080
make verify
make proto
docker compose config
docker compose up --build
```

Individual builds:

```powershell
go build -o bin/ft12-client ./cmd/ft12-client
go build -o bin/ft12-emulator ./cmd/ft12-emulator
go build -o bin/ft12-gateway ./cmd/ft12-gateway
go build -o bin/ft12-cli ./cmd/ft12-cli
go build -o bin/ft12-api ./cmd/ft12-api
```

The `Makefile` exposes the same common operations for environments with `make`.
`make verify` runs formatting, architecture checks, tests, and build.
`make clean-runtime-dry-run` previews ignored runtime/build cleanup, and
`make clean-runtime` removes those ignored local artifacts.

Useful service runs:

```powershell
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -mode crc16
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -mode crc16 -schedule interval -interval 1s
```

gRPC control:

```powershell
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -grpc-listen 127.0.0.1:9100
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -grpc-listen 127.0.0.1:9200
```

Docker Compose:

```powershell
docker compose up --build
docker compose down -v
```

The API is available on `http://localhost:8080` and serves `/api`, `/health`,
and `/metrics`.

## Active And Legacy Code

Active Go code lives in `cmd/` and `internal/`.

Legacy/reference code lives in:

- `legacy/python`;
- `legacy/go_sln`.

Legacy code is preserved for comparison and is not part of normal root module
build/test workflows.

Go formatting checks apply to active Go files and intentionally exclude
`legacy/`. The repository includes `.gitattributes` so active text files use LF
in Git. The dedicated Go format scripts compare `gofmt` output after CRLF/LF
normalization, which keeps Windows CI stable without weakening active-code
formatting.

## Local Logs And Cleanup

Runtime logs default to `runtime/logs`:

- `ft12-emulator.log`
- `ft12-gateway.log`
- `ft12-api.log`

The `-log` flag overrides the path; `-log=` sends logs to stdout. Do not log
secrets or request bodies. Protocol frame hex is logged for local diagnostics.

Cleanup scripts remove ignored runtime/build artifacts only:

```sh
go run ./tools/checks clean-runtime --dry-run
go run ./tools/checks clean-runtime
```

## Logs And Reports

Everything a run produces goes under `runtime/`, which is gitignored and
removed by `go run ./tools/checks clean-runtime`:

```text
runtime/logs/       service logs, when a service is given a log file
runtime/reports/    coverage and test output for both stacks
```

```sh
make reports
```

Writes `go-coverage.out` and an HTML view of it, the same for the console
under `web-coverage/`, and the machine-readable run of each suite. The Go
profile is measured with `-coverpkg=./...`, so a package counts as covered by
whichever test exercises it rather than only by tests living beside it.

```sh
go run ./tools/checks coverage --min 70
```

Summarises the profile and fails below a floor. It reports hand-written code
separately from the total: the protobuf output is nearly a third of the
statements in this module, and a number that moves when the schema changes
rather than when the tests do is measuring the wrong thing.

## Repository Checks

The rules the compiler and the test suite cannot express live in
`tools/checks`, one Go command with a subcommand each:

```sh
go run ./tools/checks architecture    # dependency boundaries
go run ./tools/checks format          # gofmt over tracked Go files
go run ./tools/checks doc-links       # local Markdown links resolve
go run ./tools/checks clean-runtime   # remove build and runtime artefacts
go run ./tools/checks release         # all of the above, plus tests and build
```

The architecture check keeps `internal/protocol` independent of transports,
config, logging, the service packages, the gRPC adapters and any future adapter
layer, and keeps the emulator and the gateway independent of each other. Active
code must not import `legacy/`.

It reads the real import graph from `go list`, so it sees a forbidden package
reached through two hops as clearly as a direct import, and reports the chain
and the line the first hop is written on. Production code is checked
transitively; test files are checked for direct imports only.

## Local CI Flow

When all tools are available, use:

```powershell
go fmt ./...
.\scripts\check-go-format.ps1
.\scripts\check-architecture.ps1
go test ./...
go build ./...
docker compose config
docker compose build
.\scripts\check-doc-links.ps1
.\scripts\release-check.ps1
```
