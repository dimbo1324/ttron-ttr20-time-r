# Frame Dumps

The active wire format is:

```text
0x68 | LEN | 0x68 | CONTROL | ADDRESS | DATA... | CHECKSUM | 0x16
```

## Read-Time Request, sum mode

```text
68 03 68 00 01 01 02 16
```

Checksum details:

```text
00 + 01 + 01 = 02
```

## Read-Time Request, crc16 mode

```text
68 03 68 00 01 01 90 20 16
```

CRC details:

```text
CRC16/Modbus over 00 01 01 = 0x2090
wire order = 90 20
```

## Example Read-Time Response Payload

Timestamp payload for `2026-06-04 15:04:05`:

```text
01 32 30 32 36 2d 30 36 2d 30 34 20 31 35 3a 30 34 3a 30 35
```

The first byte is command `0x01`; the remaining bytes are ASCII timestamp text.
