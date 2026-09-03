# Emulator Service

A TCP device. It accepts connections, parses FT1.2-like frames with the
protocol core, answers read-time and read-identity, and behaves badly on
request.

```sh
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000
```

CRC-16/Modbus instead of the additive checksum:

```sh
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -mode crc16
```

The gateway has to be told the same thing, or nothing it sends will decode.

## Fault modes

Each one exercises a specific part of the gateway rather than merely breaking
things:

| Flag | What the device does | What it tests |
| --- | --- | --- |
| `-delay <ms>` | answers late | request timeouts |
| `-bad-checksum` | corrupts every response checksum | in-session retries |
| `-badcrc <0..1>` | corrupts that share of them | retries under intermittent noise |
| `-fragment-response` | writes the answer in pieces | the streaming parser |
| `-fragment <0..1>` | fragments that share of them | the parser, intermittently |
| `-fragment-delay` | how far apart the pieces are | the parser's patience |
| `-no-response` | says nothing at all | timeout, then availability |
| `-close-after-request` | drops the connection immediately | reconnect and backoff |

A flag with a probability and a flag without are two different statements. The
probability forms are the interesting ones — a line that fails one frame in
three is harder to handle correctly than one that fails every frame, and it is
what a real line does.

```sh
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -badcrc 0.35
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -delay 2000
go run ./cmd/ft12-emulator -listen 127.0.0.1:9000 -no-response
```

All of it is also settable at run time, over the API and from the console —
see [HTTP API](http-api.md) and [the tour](tour.md).

## The device clock

```sh
go run ./cmd/ft12-emulator -clock-offset 5s -clock-drift 2m
```

`-clock-offset` shifts the time the device reports by a constant;
`-clock-drift` moves it by that much per day. Together they are what the
gateway's skew monitoring exists to notice, and the one fault where the
protocol works flawlessly and the data is wrong anyway.

## Identity

```sh
go run ./cmd/ft12-emulator -identity-model TTR20 -identity-serial SN-0000042 -identity-firmware 1.2.3
```

The gateway probes read-identity once per connection and remembers a device
that does not support it, so an unsupported device costs one request per
connection and never fails a poll.

## Control plane

With `-grpc-listen` set (the default is `:9100`), the emulator exposes a gRPC
control API: status, the current fault mode, a new fault mode, and the tail of
its event ring. It keeps in-memory counters and a fixed-size recent event
buffer, and nothing survives a restart.

See [gRPC API](grpc-api.md).
