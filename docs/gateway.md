# Gateway Service

The gateway is a TCP polling service. It connects to an FT1.2-like
emulator/device, sends read-time requests, parses responses, tracks status, and
reconnects with backoff after errors.

## Schedule

With no flags the gateway reads the device clock **on the fifth second of
every minute** -- the brief this project answers. That is an *aligned*
schedule: poll instants are measured from the epoch, so they land on the same
wall-clock boundary after every restart and reconnect rather than drifting by
however long each reply took.

```powershell
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -mode sum
go run ./cmd/ft12-gateway -target 127.0.0.1:9000
```

For a fixed rate -- every N seconds from the last poll, phase wherever the
connection put it -- ask for it by name:

```powershell
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -schedule interval -interval 5s
```

An aligned schedule needs its offset to fall inside its interval, so
`-interval` on its own is refused while the default `+5s` offset is still in
force. The error names both flags.

CRC16:

```powershell
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -mode crc16
go run ./cmd/ft12-gateway -target 127.0.0.1:9000 -mode crc16
```

Useful flags:

- `-target`: TCP address of the emulator/device;
- `-schedule`: `aligned` (default) or `interval`;
- `-interval`: polling interval;
- `-poll-offset`: offset inside an aligned interval;
- `-timeout`: request/response timeout;
- `-connect-timeout`: TCP connect timeout;
- `-backoff-initial`, `-backoff-max`: reconnect behavior;
- `-recent`: in-memory recent event buffer size.

## Device Inventory

Given `-devices <file>`, the gateway polls every enabled device in a JSON
inventory instead of the single `-target`. Each entry carries its own schedule,
timeouts, clock thresholds and availability policy, so one gateway can run a
minute-aligned meter next to a five-second one.

See `examples/devices.example.json` for the full field list and
`examples/devices.demo.json` for a three-device set — one of them deliberately
pointed at a dead port — that gives the console's fleet view something to show.

Without an inventory the gateway still answers the fleet question, reporting a
fleet of one, so a consumer never has to ask which mode it is in.

## Control Plane

The gateway service also exposes a gRPC control API. The HTTP API
use that control plane for status, recent events, and polling start/stop
operations. Persistence remains future work.

Beyond the connection counters, the status it reports carries the measured
clock (skew, median, drift and the fit of that drift), device health with its
hysteresis thresholds and latency percentiles, the poll schedule and the retry
budget. `GetHistory` returns the two rolling windows those aggregates are
computed from, and `UpdateSettings` changes the schedule, retry budget, clock
thresholds and availability policy of a gateway that is already polling. See
[HTTP API](http-api.md) for the JSON shapes.

In inventory mode the control plane is bound to the primary device -- the first
by id -- so starting, stopping and reconfiguring act on that one device rather
than on the fleet.
