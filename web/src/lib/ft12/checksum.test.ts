import {
  CHECKSUM_MODES,
  checksumLength,
  checksumMatches,
  computeChecksum,
  crc16,
  isChecksumMode,
  sum8,
} from "./checksum";

/**
 * These are the vectors the Go implementation is tested against
 * (`internal/protocol/checksum`). Sharing them is the only thing keeping two
 * implementations of one wire format honest.
 */
describe("sum8", () => {
  it.each([
    { name: "empty", data: [], want: 0x00 },
    { name: "read-time request body", data: [0x00, 0x01, 0x01], want: 0x02 },
    { name: "wraps at 256", data: [0xff, 0x02], want: 0x01 },
    { name: "masks values above a byte", data: [0x1ff], want: 0xff },
  ])("$name", ({ data, want }) => {
    expect(sum8(data)).toBe(want);
  });
});

describe("crc16", () => {
  it("matches the CRC-16/Modbus check value", () => {
    expect(crc16([..."123456789"].map((char) => char.charCodeAt(0)))).toBe(0x4b37);
  });

  it.each([
    { name: "empty", data: [], want: 0xffff },
    { name: "single zero byte", data: [0x00], want: 0x40bf },
  ])("$name", ({ data, want }) => {
    expect(crc16(data)).toBe(want);
  });

  it("stays inside 16 bits", () => {
    const value = crc16([0xff, 0xff, 0xff, 0xff]);
    expect(value).toBeGreaterThanOrEqual(0);
    expect(value).toBeLessThanOrEqual(0xffff);
  });
});

describe("checksumLength", () => {
  it.each([
    { mode: "sum" as const, want: 1 },
    { mode: "crc16" as const, want: 2 },
  ])("$mode is $want byte(s)", ({ mode, want }) => {
    expect(checksumLength(mode)).toBe(want);
  });
});

describe("isChecksumMode", () => {
  it.each(CHECKSUM_MODES)("accepts %s", (mode) => {
    expect(isChecksumMode(mode)).toBe(true);
  });

  it.each(["md5", "", "SUM", "crc32"])("rejects %s", (value) => {
    expect(isChecksumMode(value)).toBe(false);
  });
});

describe("computeChecksum", () => {
  it("returns one byte in sum mode", () => {
    expect(computeChecksum("sum", [0x00, 0x01, 0x01])).toEqual([0x02]);
  });

  it("returns crc16 little-endian", () => {
    const payload = [0x00, 0x01, 0x01];
    const value = crc16(payload);
    expect(computeChecksum("crc16", payload)).toEqual([value & 0xff, (value >> 8) & 0xff]);
  });
});

describe("checksumMatches", () => {
  const payload = [0x00, 0x01, 0x01];

  it.each(CHECKSUM_MODES)("accepts the computed checksum in %s mode", (mode) => {
    expect(checksumMatches(mode, payload, computeChecksum(mode, payload))).toBe(true);
  });

  it("rejects a corrupted checksum", () => {
    const actual = computeChecksum("sum", payload);
    expect(checksumMatches("sum", payload, [actual[0]! ^ 0xff])).toBe(false);
  });

  it("rejects a checksum of the wrong length", () => {
    expect(checksumMatches("crc16", payload, [0x02])).toBe(false);
    expect(checksumMatches("sum", payload, [0x02, 0x00])).toBe(false);
  });
});
