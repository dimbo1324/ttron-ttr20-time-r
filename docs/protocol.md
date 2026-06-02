# Protocol

## Current Frame Format

The active Go baseline uses a simplified FT1.2-like variable frame:

```text
0x68 | LEN | 0x68 | CONTROL | ADDR | DATA... | CHECKSUM | 0x16
```

`LEN` is the length of `CONTROL + ADDR + DATA`.

The project currently supports two checksum modes:

- `sum`: 8-bit additive checksum over `CONTROL + ADDR + DATA`;
- `crc16`: CRC-16/Modbus over `CONTROL + ADDR + DATA`, little-endian on the wire.

## Read-Time Command

The client sends command `0x01` in `DATA`.

Example request with `sum`:

```text
68 03 68 00 01 01 02 16
```

The emulator responds with:

```text
DATA = 0x01 + ASCII("YYYY-MM-DD HH:MM:SS")
```

The response control byte currently mirrors the request control byte with bit
`0x80` set.

## Known Simplifications

This is not yet a full protocol core. Step 2 will define stronger frame models,
validation behavior, command types, test vectors, and clearer separation between
wire codec and application commands.

The original task document and retained PDF live under `task/` and `docs/files/`.
