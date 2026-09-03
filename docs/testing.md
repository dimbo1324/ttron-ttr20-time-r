# Testing

469 Go tests across 69 files, and 33 Jest suites for the console. Both run on
every push; see [CI](ci.md).

## Running them

```sh
go test ./...
go test -race ./...
pnpm --dir web test
```

Everything a release gate checks, in one command:

```sh
go run ./tools/checks release
```

## What is covered

**The protocol core.** Additive checksum and CRC-16/Modbus, including vectors
and mode-parsing edges. Frame encode and decode, invalid lengths, truncated
checksums, and every typed validation error. The streaming parser against
fragmentation, noise before the start byte, several frames in one chunk,
recovery after an invalid frame, and the maximum-size guard. Read-time and
read-identity request and response parsing, including timestamps of the wrong
length and the wrong shape. Codec helpers, including a response carrying the
wrong command.

**The gateway's reasoning**, each package against a table rather than a live
device: schedule instants for both modes across restarts, the retry policy's
budget and backoff, skew classification on a median with drift and its fit,
health hysteresis and latency percentiles, and the device inventory's parsing
and defaults.

**The services.** Emulator behaviour under every fault mode, its history and
its status counters. Gateway polling, timeouts, backoff, reconnects, history
and counters. A `-race` test that hammers `UpdateSettings` from four
goroutines while the gateway is polling, because settings that can be changed
mid-poll are exactly the kind of thing that is fine until it is not.

**The adapters.** Both gRPC control APIs. The shared mapper. HTTP DTO
nil-safety, UTC timestamp formatting, and status, checksum and direction
mapping. Handlers, the error envelope, CORS, readiness, metrics, the merged
events feed, controls, security headers, JSON body limits, malformed JSON and
invalid limits. The export endpoints: download envelopes, CSV escaping, source
validation, method validation and limit validation.

**Everything else that has been wrong once.** Config loading and validation.
Lifecycle cancellation and error propagation. Stable event-ring IDs and
snapshot copying. Runtime logging helpers and the cleanup dry-run. Hex dump
formatting. The repository's own checks, including the coverage profile merge
— a bug that made a well-tested module report six percent.

**The console.** The browser-side protocol implementation against the same
vectors the Go core is tested with, the bench engine, the telemetry seam from
both sources, the API client and its zod parsers, the `/upstream` proxy route
(address resolution, query strings, methods, hop-by-hop headers, and the bare
502 that a dead upstream has to produce), the stores, and the components.

## Coverage

```sh
make reports
go run ./tools/checks coverage --min 70
```

A little over **77%** of hand-written Go statements at the time of writing,
against a floor of 70% enforced in CI. Generated protobuf is counted
separately: it is nearly a quarter of the statements in this module, and
folding it in produces a number that moves when the schema changes rather than
when the tests do.

The floor is a floor rather than a ratchet. A number that must never fall
turns every refactor into a negotiation with the metric; a floor says only
that the suite has not quietly stopped covering the code.

The profile is measured with `-coverpkg=./...` so that a package counts as
covered by whichever test exercises it — without it, `internal/api/http/metrics`
read 0% while being exercised on every request in the handler tests. That flag
has a cost the summary has to undo: every test binary then reports on every
package, so each span appears once per binary, and summing them turns 4,700
statements into 122,000. The coverage subcommand folds repeated spans, keeping
the highest count, before it counts anything.

## Docker

```sh
docker compose config
docker compose build
docker compose up -d
```

```sh
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/api/v1/ready
curl http://127.0.0.1:8080/api/v1/overview
curl http://127.0.0.1:8080/api/v1/events
curl http://127.0.0.1:8080/api/v1/gateway/fleet
curl http://127.0.0.1:8080/metrics
curl http://127.0.0.1:3000/en
curl http://127.0.0.1:3000/upstream/api/v1/gateway/fleet
```

The last one matters more than it looks: it goes through the console's own
proxy, which is the only check that `FT12_API_URL` is read at run time rather
than baked into the image at build time.

```sh
docker compose down -v
```

## Manual smoke

Worth doing before a release, in a browser:

- client and emulator in `sum` and in `crc16`; gateway and emulator likewise;
- every fault mode switched on from the console and its effect visible in the
  monitor;
- the bench and the live source showing the same page with different numbers;
- the live source with the API stopped — the console must name the missing
  processes, not render zeros;
- RU and EN, and the reference page in both;
- no text overflow, no body-level horizontal scroll, and no overlapping
  elements on the dashboard, monitor and gateway views.

## Fuzzing

`make test-fuzz` documents the current entrypoint status. Fuzzing is not part
of the gate because no stable corpus is configured for this milestone; the
frame decoder and the streaming parser are the two places it would pay.
