# HTTP API

Step 6 adds a thin HTTP/JSON adapter for browser and local tooling access.
Business logic remains in emulator and gateway services behind the existing
gRPC control plane.

## Run

```powershell
go run ./cmd/ft12-api -http-listen 127.0.0.1:8080 -emulator-grpc 127.0.0.1:9100 -gateway-grpc 127.0.0.1:9200
```

## Endpoints

- `GET /health`
- `GET /api/v1/health`
- `GET /api/v1/ready`
- `GET /metrics`
- `GET /api/v1/config`
- `GET /api/v1/overview`
- `GET /api/v1/emulator/status`
- `GET /api/v1/emulator/fault-mode`
- `PUT /api/v1/emulator/fault-mode`
- `GET /api/v1/emulator/events?limit=100`
- `GET /api/v1/gateway/status`
- `POST /api/v1/gateway/start`
- `POST /api/v1/gateway/stop`
- `GET /api/v1/gateway/last-read-time`
- `GET /api/v1/gateway/events?limit=100`
- `GET /api/v1/gateway/fleet`
- `GET /api/v1/gateway/history`
- `PUT /api/v1/gateway/settings`
- `GET /api/v1/events?source=all&limit=100`
- `GET /api/v1/export/events.json?source=all&limit=100`
- `GET /api/v1/export/events.csv?source=all&limit=100`
- `GET /api/v1/export/overview.json?limit=50`
- `GET /api/v1/export/emulator-status.json`
- `GET /api/v1/export/gateway-status.json`

## Gateway Status

`GET /api/v1/gateway/status` returns the flat connection counters plus five
nested sections that describe how the gateway is behaving, not just whether it
is up:

- `schedule` — poll timing: `mode` (`interval` or `aligned`), `intervalMs`,
  `offsetMs`, and `nextPollAt`.
- `retry` — the in-session retry budget (`attempts`, `delayMs`, `maxDelayMs`)
  and how much of it has been spent (`totalRetries`, `exhaustedPolls`).
- `clock` — the measured device clock. `skewMs` is signed: positive means the
  device reads ahead of the gateway. `medianSkewMs` is what `state`
  (`unknown` / `ok` / `warn` / `critical`) is judged on, because a single
  sample carries the whole round trip. `driftPerDayMs` is a least-squares rate
  over the sample window and `driftFit` is its R², so a rate with a poor fit
  can be recognised as noise.
- `health` — reachability with hysteresis (`state`, `availability`,
  `consecutiveFailures`, the `degradeAfter` / `offlineAfter` / `recoverAfter`
  policy) and the latency percentiles the decision was made on.
- `identity` — the nameplate the read-identity probe found. `supported` goes
  false for a device that answers the probe with a plain acknowledgement.

Every duration in these sections is milliseconds. `clock.state` and
`health.state` are always a named state, never an empty string.

## Gateway Fleet

`GET /api/v1/gateway/fleet` returns `summary` — device counts by health and
clock state, plus `worstClockSkewMs` and `worstClockDeviceId` — and `devices`,
an array of the same status objects.

A gateway started without a device inventory reports a fleet of one, so a
consumer renders the same view in either configuration and never has to ask
which mode the gateway is in.

## Gateway Settings

`PUT /api/v1/gateway/settings` reconfigures a gateway that may be mid-poll:

```json
{
  "scheduleMode": "aligned",
  "pollIntervalMs": 60000,
  "pollOffsetMs": 5000,
  "requestTimeoutMs": 1500,
  "retryAttempts": 2,
  "retryDelayMs": 200,
  "clockWarnMs": 2000,
  "clockCriticalMs": 30000,
  "degradeAfter": 3,
  "offlineAfter": 10,
  "recoverAfter": 2
}
```

The reply carries `settings` -- what was actually applied, which is not always
what was asked for, since the retry policy fills in its own maximum backoff --
and `status`, so a consumer redraws without a second round trip.

The whole configuration is sent, not a patch. A partial update needs a field
mask to tell "leave this alone" apart from "set this to zero", and on a control
plane with one operator that machinery is not repaid.

Three properties worth relying on:

- **All or nothing.** Everything is validated before anything is applied, so a
  rejected request leaves the gateway exactly as it was.
- **Nothing in flight is interrupted.** The request timeout a poll started
  under is the one it finishes under.
- **A schedule change is not queued behind the old interval.** A gateway
  parked on a one-minute wait re-plans immediately when told to poll every
  second.

Rejected values return `400 INVALID_ARGUMENT` with the reason in the message --
for example a request timeout at or above the poll interval, an offset at or
above the interval, or a critical clock threshold below the warning one.

**What is not settable, and why.** The target address, checksum mode and
adapter address are identity rather than settings: change one mid-flight and
the frames on the wire stop matching the device. The clock and health window
sizes are absent because resizing a ring buffer discards the history in it.

**Which device.** In inventory mode the control plane is bound to the primary
device (the first by id), so this reconfigures that one and not the fleet.
`status.deviceId` names it.

## Gateway History

`GET /api/v1/gateway/history` returns the two rolling windows the aggregates
are computed from, oldest first:

- `clockSamples` — `at`, `skewMs`, `roundTripMs` per measurement.
- `healthOutcomes` — `at`, `success`, `latencyMs` per poll. `latencyMs` is
  meaningful only where `success` is true.

It is a separate call from `gateway/status` so an ordinary status read stays a
fixed size regardless of the window. A consumer that only needs the current
numbers never pays for the history.

## Fault Mode Update

```json
{
  "responseDelayMs": 0,
  "corruptChecksum": false,
  "corruptChecksumProbability": 0,
  "fragmentResponse": false,
  "fragmentProbability": 0,
  "fragmentDelayMs": 40,
  "noResponse": false,
  "closeAfterRequest": false
}
```

## Error Model

Errors are returned consistently:

```json
{
  "error": {
    "code": "GATEWAY_UNAVAILABLE",
    "message": "gateway gRPC service is unavailable"
  }
}
```

The adapter maps malformed JSON and validation errors to `400`, unsupported
methods to `405`, upstream gRPC failures to `502`/`503`, and deadlines to `504`.

## Readiness

`GET /api/v1/ready` checks that the API process can reach both upstream gRPC
services. It returns `200` when emulator and gateway status calls succeed, and
`503` when either upstream is unavailable.

## Metrics

`GET /metrics` returns Prometheus-compatible text metrics for HTTP request
counts and total request duration by method, path, and status.

`GET /health` also includes build metadata fields: `version`, `commit`, and
`buildDate`.

## Exports

Export endpoints are read-only and use the same upstream gRPC status/events
data as the ordinary JSON API. They do not read local log files and do not
accept filesystem paths.

`source` supports `all`, `emulator`, or `gateway`. `limit` must be an integer
from `1` to `1000`; invalid values return `400 INVALID_LIMIT`.

Events CSV columns:

```text
timestamp,source,service,direction,command,checksumMode,remoteAddr,rawHex,message,error
```

CSV output is generated with Go's standard CSV writer, so commas, quotes, and
line breaks are escaped correctly. JSON downloads are indented and include an
`exportedAt` timestamp. Responses set `Content-Type` and `Content-Disposition`
headers with server-generated filenames such as
`ft12-events-YYYYMMDD-HHMMSS.csv`.

Exported files may contain protocol diagnostic data, raw hex, remote addresses,
and service counters. Treat them as local troubleshooting artifacts.

## Architecture

HTTP handlers depend on small interfaces implemented by gRPC client adapters.
They do not import emulator or gateway service packages directly, and they do
not duplicate protocol or polling behavior.
