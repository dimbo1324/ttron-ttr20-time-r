# Protocol

## Status

Step 2 adds a typed FT1.2-like protocol core under `internal/protocol`.

The core is intentionally independent from TCP, gRPC, Web UI, Docker, config,
logging, and emulator/gateway service logic. It uses the Go standard library
only.

## Package Layout

```text
internal/protocol/
  checksum/   checksum modes, sum8, CRC16, compute/verify helpers
  frame/      frame model, encoder, decoder, typed errors, stream parser
  command/    command IDs and read-time request/response payloads
  codec/      high-level read-time frame helpers
```

## Frame Format

The active wire format is a simplified FT1.2-like variable-length frame:

```text
0x68 | LEN | 0x68 | CONTROL | ADDRESS | DATA... | CHECKSUM | 0x16
```

Byte semantics:

- `0x68`: start byte;
- `LEN`: length of `CONTROL + ADDRESS + DATA`;
- repeated `0x68`: variable-length frame marker;
- `CONTROL`: one protocol control byte;
- `ADDRESS`: one adapter/device address byte;
- `DATA`: zero or more command payload bytes;
- `CHECKSUM`: one `sum` byte or two `crc16` bytes;
- `0x16`: end byte.

The current Go client sends request control `0x00` and the emulator returns
responses with bit `0x80` set on the request control byte.

## Checksum Modes

`sum`:

- one byte;
- additive checksum modulo 256;
- computed over `CONTROL + ADDRESS + DATA`.

`crc16`:

- two bytes;
- CRC-16/Modbus algorithm, polynomial `0xA001`, initial value `0xFFFF`;
- little-endian on the wire;
- computed over `CONTROL + ADDRESS + DATA`.

Decoding is mode-aware because checksum length changes the expected frame size.

## Read-Time Command

Command ID:

```text
0x01
```

Read-time request payload:

```text
DATA = 0x01
```

Example request in `sum` mode:

```text
68 03 68 00 01 01 02 16
```

Read-time response payload:

```text
DATA = 0x01 + ASCII("YYYY-MM-DD HH:MM:SS")
```

Timestamp layout:

```text
2006-01-02 15:04:05
```

The command package validates command IDs, timestamp length, and timestamp
format. Parsed timestamps use Go's `time.Parse` with the fixed layout above.

## Streaming Parser

TCP is a byte stream, so the frame package includes a `StreamParser`.

The parser:

- accepts arbitrary byte chunks;
- retains partial frames between pushes;
- emits complete decoded frames;
- handles fragmented frames;
- handles multiple frames in one chunk;
- discards noise before the start byte;
- reports protocol errors and resynchronizes where possible;
- protects against unbounded buffer growth with a max frame size.

Client and emulator runtime code now use this parser instead of assuming one
network read equals one complete frame.

## Error Handling

The frame and command packages expose sentinel errors that work with
`errors.Is`, including:

- invalid checksum mode;
- frame too short;
- invalid start byte;
- invalid repeated start byte;
- invalid length;
- invalid checksum;
- invalid end byte;
- frame too large;
- empty command payload;
- unexpected command;
- invalid read-time payload;
- invalid timestamp.

Runtime code logs protocol errors without panicking.

## Known Simplifications

This is a simulation/learning protocol core, not a certified industrial
implementation. The current frame format is FT1.2-like and intentionally narrow:

- only the read-time command is modeled as a first-class command;
- real TTR20/FT1.2 device behavior may require additional frame variants,
  address handling, control semantics, timing rules, and certification work;
- serial transport, security, and expanded command coverage remain future work.
  gRPC, gateway polling, Web UI, Docker, CI, and observability are implemented
  around this protocol core without changing its package boundaries.

The original task document and retained protocol PDF live under `task/` and
`docs/files/`.
