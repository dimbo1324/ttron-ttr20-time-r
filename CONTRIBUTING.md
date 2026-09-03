# Contributing

Thanks for looking at this project. It is structured as a small industrial
protocol simulation platform, and the thing most worth preserving in a change
is the separation between the protocol core, the service runtime, the API
adapters and the deployment tooling.

## Prerequisites

- Go, the version declared in `go.mod`.
- Node 22 for the console (`web/package.json` sets the floor) and pnpm 9, the
  version CI installs.
- Docker, for the Compose checks. Everything else works without it.

`make` is convenient and never required: the repository's checks are one Go
program.

## Setup

```sh
git clone https://github.com/dimbo1324/ttron-ttr20-time-r.git
cd ttron-ttr20-time-r
go test ./...
pnpm --dir web install
```

## Before opening a pull request

```sh
go run ./tools/checks release
```

Formatting, dependency boundaries, documentation links, `go vet`, the tests,
the build, Compose configuration and a cleanup dry-run. Then the things it
does not cover:

```sh
make lint
go test -race ./...
pnpm --dir web lint
pnpm --dir web typecheck
pnpm --dir web test
```

CI runs the same commands. If one of them fails here it will fail there.

## Architecture rules

These are enforced by `go run ./tools/checks architecture`, not merely
requested:

- `internal/protocol` depends on the standard library and nothing else — not
  TCP, gRPC, HTTP, config, logging, Docker, or the service packages.
- `internal/emulator` and `internal/gateway` do not import each other.
- `internal/api/http` does not import the service packages; it talks to them
  through the gRPC clients.
- Nothing active imports `legacy/`.

Beyond that: do not duplicate protocol logic in an adapter, do not edit
generated gRPC files by hand, and do not change the wire format or the gRPC
contracts without a deliberate design change and tests to go with it.

## Protobuf

Regenerate only when a proto source has intentionally changed:

```sh
make proto
```

If `make` is unavailable, the `protoc` invocation is written out in the
`Makefile`.

## Documentation

Documentation lives in `docs/` and is checked: `go run ./tools/checks
doc-links` fails on a local Markdown link that does not resolve, and CI runs
it.

The screenshots in [docs/tour.md](docs/tour.md) are produced by a script
against a running stack rather than taken by hand. If a change alters the
console's interface, re-run it — see
[tools/docmedia](tools/docmedia/README.md).

## Generated artefacts

Do not commit `bin/`, `dist/`, `tmp/`, `runtime/`, logs or local `.exe` files.
Everything a run produces belongs under `runtime/`, which is gitignored.

```sh
go run ./tools/checks clean-runtime --dry-run
go run ./tools/checks clean-runtime
```

## Branches and commits

Focused branches, and commit subjects in the imperative that say what changed
and why it mattered:

- `fix: answer an upstream timeout with 504 rather than 502`
- `docs: bring the CI job table in line with the workflow`
- `test: cover crc16 parser resynchronisation after noise`
