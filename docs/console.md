# Web Console

The console under `web/` is a Next.js app that shows what the protocol is
doing: frame-by-frame decoding, clock skew and drift, device availability, and
the polling schedule. Its interface is Russian by default with English as a
second locale (`/ru`, `/en`).

It reads from one of two sources, chosen by the switch in the header.

## Bench

The default. A working model of the device, the line and the gateway, running
entirely in the browser — the same FT1.2 encoding, the same checksum
algorithms, the same skew and hysteresis rules as the Go services. Nothing
needs to be installed or running.

This is the teaching mode: every fault can be switched on and its effect
watched on the timeline, and a frame copied out of the log decodes identically
in the analyzer.

```powershell
pnpm --dir web dev
```

## Live stack

The same console, reading a running Go stack over the HTTP API. Every number on
screen then comes from the real gateway and the real device.

Start three processes, each in its own terminal:

```powershell
go run ./cmd/ft12-emulator -grpc-listen 127.0.0.1:9100
go run ./cmd/ft12-gateway -grpc-listen 127.0.0.1:9200 -devices examples/devices.demo.json
go run ./cmd/ft12-api -http-listen 127.0.0.1:8080 -emulator-grpc 127.0.0.1:9100 -gateway-grpc 127.0.0.1:9200
```

Then switch the header control to **Live stack**. The choice is remembered in
the browser.

`examples/devices.demo.json` runs three devices, one of them pointed at a port
with nothing listening, so the fleet table has something to show: two devices
answering and one going degraded and then offline.

The gateway can also run without an inventory, against a single `-target`. The
console renders the same view either way — a gateway without an inventory
reports a fleet of one.

### How the requests get there

The browser only ever talks to the console's own origin. `next.config.ts`
rewrites `/upstream/*` to the API, whose address comes from `FT12_API_URL`
(default `http://127.0.0.1:8080`). That keeps the API off the public network in
a deployment, removes CORS from the picture, and means the page never needs to
know where the API lives.

### What the live source can and cannot change

It can start and stop polling, and it can set the emulator's fault mode — delay,
corrupt checksum, fragmentation, silence, and dropping the connection. Those
drive the real emulator, which is what makes this a bench rather than a viewer.

It cannot change the gateway's interval, offset, timeouts, clock thresholds or
availability policy: those are the gateway process's own configuration and its
control plane has no setter for them. The console shows the running values and
says they are read-only rather than offering a control that would discard the
change.

The device clock offset and drift are bench-only. The Go emulator answers with
its host's real clock and has no offset to inject, so those two sliders are
disabled on the live source and say why.

### When the API is not running

The console says so, names the processes that are missing, and shows the
commands to start them. It does not fall back to rendering zeros — a page full
of zeros is indistinguishable from a healthy, idle gateway, which is the one
mistake a monitoring view must not make.

## Tests

```powershell
pnpm --dir web test
pnpm --dir web typecheck
```
