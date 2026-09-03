# Development

## Toolchain

- Go `1.26` (the exact version is in `go.mod`).
- Node 22 for the console (`web/package.json` sets the floor), and pnpm 9,
  which is the version CI installs.
- Docker for the Compose stack; every check that needs it says so and skips
  when it is absent.

The repository is one root Go module on purpose. `go.work` is not used.

## The short version

```sh
go run ./tools/checks release
```

Formatting, dependency boundaries, documentation links, `go vet`, the test
suite, the build, Compose configuration and a cleanup dry-run — cheapest
first, so a missing `gofmt` is a five-second answer rather than one that
arrives after a Compose build.

With `make`, the same thing is `make release-check`; `make verify` is the
subset that does not need Docker.

## Running the services

```sh
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -grpc-listen 127.0.0.1:9100
go run ./cmd/ft12-gateway  -target 127.0.0.1:9000 -grpc-listen 127.0.0.1:9200
go run ./cmd/ft12-api      -http-listen 127.0.0.1:8080 -emulator-grpc 127.0.0.1:9100 -gateway-grpc 127.0.0.1:9200
go run ./cmd/ft12-client   -host 127.0.0.1 -port 9000 -crc sum
```

With no flags at all, the gateway does what the brief asks: reads the device
clock on the fifth second of every minute. For a bench you are watching, ask
for a fixed rate instead:

```sh
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -schedule interval -interval 1s
```

CRC-16/Modbus instead of the additive checksum, on both ends:

```sh
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -mode crc16
go run ./cmd/ft12-gateway  -target 127.0.0.1:9000 -mode crc16
```

A fleet rather than one device:

```sh
go run ./cmd/ft12-gateway -devices examples/devices.demo.json
```

## The console

```sh
pnpm --dir web install
pnpm --dir web dev        # http://localhost:3000/en
pnpm --dir web test
pnpm --dir web typecheck
pnpm --dir web lint
pnpm --dir web build
```

`make lint-web` runs ESLint through the same path CI does.

## Everything at once

```sh
docker compose up --build
docker compose down -v
```

The console lands on `http://localhost:3000/en` and the API on
`http://localhost:8080`. For hot reload against a bind-mounted source tree,
see the development overlay in [Docker](docker.md).

## Individual builds

```sh
go build -o bin/ft12-emulator    ./cmd/ft12-emulator
go build -o bin/ft12-gateway     ./cmd/ft12-gateway
go build -o bin/ft12-api         ./cmd/ft12-api
go build -o bin/ft12-client      ./cmd/ft12-client
go build -o bin/ft12-cli         ./cmd/ft12-cli
go build -o bin/ft12-healthcheck ./cmd/ft12-healthcheck
```

## Repository checks

The rules the compiler and the test suite cannot express live in
`tools/checks`, one Go command with a subcommand each:

```sh
go run ./tools/checks architecture    # dependency boundaries
go run ./tools/checks format          # gofmt over tracked Go files
go run ./tools/checks doc-links       # local Markdown links resolve
go run ./tools/checks coverage        # summarise the profile, enforce a floor
go run ./tools/checks clean-runtime   # remove build and runtime artefacts
go run ./tools/checks release         # all of the above, plus tests and build
```

One Go program rather than five pairs of shell and PowerShell scripts, so
"what CI runs" and "what I can run" are the same sentence on every platform.

The architecture check keeps `internal/protocol` independent of transports,
config, logging, the service packages and the adapters, and keeps the emulator
and gateway independent of each other. It reads the real import graph from
`go list`, so it sees a forbidden package reached through two hops as clearly
as a direct import, and reports the chain and the line the first hop is
written on. Production code is checked transitively; test files are checked
for direct imports only.

## Linters

```sh
make lint        # golangci-lint
make lint-web    # ESLint
```

`golangci-lint` runs the standard set plus `bodyclose`, `errorlint`,
`unconvert`, `nilerr` and `misspell`; `errcheck` is relaxed inside `_test.go`,
where an unchecked `Close` is noise rather than a defect. The configuration is
`.golangci.yml` and needs golangci-lint v2.

The console uses ESLint's flat config (`web/eslint.config.mjs`) against
`eslint-config-next`. It is a real configuration rather than the `next lint`
that Next 16 removed: that command read "lint" as a directory name and exited
zero, which is worse than having no linter at all.

## Logs and reports

Everything a run produces goes under `runtime/`, which is gitignored and
removed by `go run ./tools/checks clean-runtime`:

```text
runtime/logs/       service logs, when a service is given a log file
runtime/reports/    coverage and test output for both stacks
```

Service logs default to `runtime/logs/<service>.log`. `-log <path>` moves one;
`-log ""` sends it to stdout, which is what the Compose stack does. Do not log
secrets or request bodies — protocol frame hex is logged, and is local
diagnostic data.

```sh
make reports
```

Writes `go-coverage.out` and an HTML view of it, the same for the console
under `web-coverage/`, and the machine-readable run of each suite.

The Go profile is measured with `-coverpkg=./...`, so a package counts as
covered by whichever test exercises it rather than only by tests living beside
it. That flag also makes every test binary report on every package, so the raw
profile names the same span once per binary; the coverage subcommand folds
those before counting.

```sh
go run ./tools/checks coverage --min 70
```

Summarises the profile and fails below a floor. It reports hand-written code
separately from the total, because the protobuf output is nearly a quarter of
the statements in this module and a number that moves when the schema changes
rather than when the tests do is measuring the wrong thing.

## Documentation media

The screenshots and animations in [the tour](tour.md) are produced by a script
against a running stack — see [tools/docmedia](../tools/docmedia/README.md).
They are not part of the ordinary build; run them after an interface change.

## Active and legacy code

Active Go code lives in `cmd/`, `internal/` and `tools/`. `legacy/python` and
`legacy/go_sln` are kept for comparison and are excluded from formatting,
build and test policy. Active code must not import them, which the
architecture check enforces.

`.gitattributes` pins LF for the text formats this repository uses. The format
check compares `gofmt` output after normalising CRLF/LF, so a Windows checkout
does not fail on line endings alone while real formatting drift still does.
