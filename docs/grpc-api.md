# gRPC API

The control plane between services. It carries status and control, never
device data: the FT1.2-like TCP path between the gateway and the device is
untouched by it.

Nothing outside the Compose network speaks gRPC. The HTTP API
([http-api.md](http-api.md)) is the outward surface, and it is a thin mapping
over this one.

## Files

```text
proto/ft12/v1/common.proto
proto/ft12/v1/emulator.proto
proto/ft12/v1/gateway.proto
```

Generated Go lands in
`github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1` and is not
edited by hand:

```sh
make proto
```

## EmulatorService

Default listen address `:9100`.

| RPC | Purpose |
| --- | --- |
| `GetStatus` | sessions, counters, last request, recent event count |
| `GetFaultMode` | the fault mode currently in force |
| `SetFaultMode` | replace it |
| `GetRecentEvents` | the tail of the event ring |

```sh
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -grpc-listen 127.0.0.1:9100
```

Turn it off with `-grpc-listen ""`.

## GatewayService

Default listen address `:9200`.

| RPC | Purpose |
| --- | --- |
| `GetStatus` | connection counters, schedule, retry budget, clock, health, identity |
| `StartPolling` / `StopPolling` | idempotent control over the same polling service |
| `GetRecentEvents` | the tail of the event ring |
| `GetLastReadTime` | the most recent successful read |
| `GetFleet` | every device this gateway polls, plus a summary |
| `GetHistory` | the two rolling windows the aggregates are computed from |
| `UpdateSettings` | reconfigure a gateway that is already polling |

```sh
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -grpc-listen 127.0.0.1:9200
```

The gateway starts polling by default.

Three of these are worth a sentence each.

**`GetFleet`** answers the same shape whether or not the gateway was given a
device inventory: without one it reports a fleet of one, so a consumer never
has to ask which mode it is in.

**`GetHistory`** is separate from `GetStatus` on purpose. The windows grow
with configuration, and a caller that only wants the current numbers should
not pay for them on every poll.

**`UpdateSettings`** takes the whole configuration rather than a patch. A
partial update needs a field mask to tell "leave this alone" apart from "set
this to zero", and on a control plane with one operator that machinery is not
repaid. It validates everything before applying anything, so a refusal leaves
the gateway exactly as it was; the JSON shape, the rules and what is
deliberately not settable are in [HTTP API](http-api.md).

In inventory mode the control plane binds to the primary device — the first by
id — so start, stop and settings act on that one device rather than on the
fleet.

## Not yet

TLS and authentication, persistence, and OpenTelemetry tracing are all absent,
deliberately, at this milestone. Do not expose these ports to an untrusted
network.
