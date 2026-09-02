/**
 * Checksum modes, mirroring `internal/protocol/checksum` on the Go side.
 *
 * The console decodes frames locally rather than asking the API to do it: the
 * frame inspector has to stay useful with no backend running (that is the
 * whole point of a teaching bench), and a byte-level view needs the checksum
 * recomputed on every keystroke, which is not a round trip worth making.
 *
 * Keeping a second implementation honest is a real cost, so both are written
 * against the same fixed vectors: `Sum8([0x00,0x01,0x01]) === 0x02` and
 * `crc16("123456789") === 0x4B37`.
 */

export const CHECKSUM_MODES = ["sum", "crc16"] as const;

export type ChecksumMode = (typeof CHECKSUM_MODES)[number];

export function isChecksumMode(value: string): value is ChecksumMode {
  return (CHECKSUM_MODES as readonly string[]).includes(value);
}

/** Bytes the checksum occupies on the wire — this is what changes frame length. */
export function checksumLength(mode: ChecksumMode): number {
  return mode === "crc16" ? 2 : 1;
}

/** Additive checksum modulo 256 over the payload. */
export function sum8(data: readonly number[]): number {
  let total = 0;
  for (const byte of data) total = (total + byte) & 0xff;
  return total;
}

/**
 * CRC-16/Modbus: reflected polynomial 0xA001, initial value 0xFFFF.
 *
 * Written as the table-free bitwise form because the table would be the only
 * thing in this module a reader could not check by eye against the spec.
 */
export function crc16(data: readonly number[]): number {
  let crc = 0xffff;
  for (const byte of data) {
    crc ^= byte & 0xff;
    for (let bit = 0; bit < 8; bit += 1) {
      crc = crc & 1 ? (crc >> 1) ^ 0xa001 : crc >> 1;
    }
  }
  return crc & 0xffff;
}

/** Checksum bytes in wire order — CRC-16 goes out little-endian. */
export function computeChecksum(mode: ChecksumMode, payload: readonly number[]): number[] {
  if (mode === "sum") return [sum8(payload)];
  const value = crc16(payload);
  return [value & 0xff, (value >> 8) & 0xff];
}

export function checksumMatches(
  mode: ChecksumMode,
  payload: readonly number[],
  actual: readonly number[],
): boolean {
  const expected = computeChecksum(mode, payload);
  return (
    expected.length === actual.length && expected.every((byte, index) => byte === actual[index])
  );
}
