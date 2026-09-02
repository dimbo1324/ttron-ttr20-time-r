import { computeChecksum, type ChecksumMode } from "./checksum";
import {
  decodeFrame,
  encodeFrame,
  END_BYTE,
  MAX_FRAME_LENGTH,
  MIN_FRAME_LENGTH,
  payloadBytes,
  scanFrames,
  START_BYTE,
  type FieldKind,
} from "./frame";

const MODES: ChecksumMode[] = ["sum", "crc16"];

/** The canonical read-time request, straight out of docs/protocol.md. */
const READ_TIME_REQUEST = [0x68, 0x03, 0x68, 0x00, 0x01, 0x01, 0x02, 0x16];

function fieldsOf(kinds: FieldKind[], raw: number[], mode: ChecksumMode) {
  const decoded = decodeFrame(raw, mode);
  return kinds.map((kind) => decoded.fields.find((field) => field.kind === kind));
}

describe("payloadBytes", () => {
  it("prefixes the data with control and address", () => {
    expect(payloadBytes(0x00, 0x01, [0x01])).toEqual([0x00, 0x01, 0x01]);
  });

  it("masks everything to a byte", () => {
    expect(payloadBytes(0x100, 0x1ff, [0x102])).toEqual([0x00, 0xff, 0x02]);
  });
});

describe("encodeFrame", () => {
  it("produces the documented read-time request", () => {
    expect(encodeFrame(0x00, 0x01, [0x01], "sum")).toEqual(READ_TIME_REQUEST);
  });

  it("counts control and address in the length byte", () => {
    const raw = encodeFrame(0x00, 0x01, [0x01, 0x02, 0x03], "sum");
    expect(raw[1]).toBe(5);
  });

  it.each(MODES)("frames in %s mode start and end correctly", (mode) => {
    const raw = encodeFrame(0x00, 0x01, [0x01], mode);
    expect(raw[0]).toBe(START_BYTE);
    expect(raw[2]).toBe(START_BYTE);
    expect(raw[raw.length - 1]).toBe(END_BYTE);
  });

  it("crc16 frames are one byte longer than sum frames", () => {
    const sum = encodeFrame(0x00, 0x01, [0x01], "sum");
    const crc = encodeFrame(0x00, 0x01, [0x01], "crc16");
    expect(crc.length).toBe(sum.length + 1);
  });
});

describe("decodeFrame", () => {
  it.each(MODES)("accepts a well-formed frame in %s mode", (mode) => {
    const raw = encodeFrame(0x00, 0x01, [0x01, 0x02], mode);
    const decoded = decodeFrame(raw, mode);

    expect(decoded.ok).toBe(true);
    expect(decoded.issues).toEqual([]);
    expect(decoded.frame).toEqual({
      control: 0x00,
      address: 0x01,
      data: [0x01, 0x02],
      raw,
    });
    expect(decoded.consumed).toBe(raw.length);
  });

  it("maps every byte to a field", () => {
    const raw = encodeFrame(0x00, 0x01, [0x01], "sum");
    const decoded = decodeFrame(raw, "sum");
    const mapped = decoded.fields.reduce((total, field) => total + field.bytes.length, 0);

    expect(mapped).toBe(raw.length);
    expect(decoded.fields.map((field) => field.kind)).toEqual([
      "start",
      "length",
      "startRepeat",
      "control",
      "address",
      "data",
      "checksum",
      "end",
    ]);
  });

  it("omits the data field when the payload is empty", () => {
    const raw = encodeFrame(0x00, 0x01, [], "sum");
    const decoded = decodeFrame(raw, "sum");

    expect(decoded.ok).toBe(true);
    expect(decoded.fields.some((field) => field.kind === "data")).toBe(false);
  });

  it("reports the checksum it expected", () => {
    const raw = encodeFrame(0x00, 0x01, [0x01], "crc16");
    const decoded = decodeFrame(raw, "crc16");

    expect(decoded.expectedChecksum).toEqual(computeChecksum("crc16", [0x00, 0x01, 0x01]));
  });

  it("marks trailing bytes beyond the frame", () => {
    const raw = [...encodeFrame(0x00, 0x01, [0x01], "sum"), 0xaa, 0xbb];
    const [trailing] = fieldsOf(["trailing"], raw, "sum");

    expect(trailing?.bytes).toEqual([0xaa, 0xbb]);
  });
});

describe("decodeFrame rejects malformed input", () => {
  it("input shorter than the minimum frame", () => {
    const decoded = decodeFrame([0x68, 0x03], "sum");

    expect(decoded.ok).toBe(false);
    expect(decoded.issues[0]?.code).toBe("tooShort");
    expect(decoded.consumed).toBe(0);
    expect(decoded.fields).toEqual([]);
  });

  it("a bad start byte", () => {
    const raw = [...READ_TIME_REQUEST];
    raw[0] = 0x00;
    const decoded = decodeFrame(raw, "sum");

    expect(decoded.ok).toBe(false);
    expect(decoded.issues.map((issue) => issue.code)).toContain("invalidStart");
    expect(decoded.issues[0]?.expected).toEqual([START_BYTE]);
    expect(decoded.issues[0]?.actual).toEqual([0x00]);
  });

  it("a bad repeated start byte", () => {
    const raw = [...READ_TIME_REQUEST];
    raw[2] = 0x00;

    expect(decodeFrame(raw, "sum").issues.map((issue) => issue.code)).toContain(
      "invalidStartRepeat",
    );
  });

  it("a payload length below control plus address", () => {
    const raw = [0x68, 0x01, 0x68, 0x00, 0x01, 0x01, 0x16];
    const decoded = decodeFrame(raw, "sum");

    expect(decoded.issues.map((issue) => issue.code)).toContain("invalidLength");
    expect(decoded.consumed).toBe(0);
  });

  it("cannot exceed the structural maximum", () => {
    // LEN is one byte, so the largest expressible frame is exactly
    // MAX_FRAME_LENGTH — a longer one is unrepresentable rather than rejected.
    const payload = new Array(0xff - 2).fill(0x01) as number[];
    const raw = encodeFrame(0x00, 0x01, payload, "crc16");

    expect(raw.length).toBe(MAX_FRAME_LENGTH);
    expect(decodeFrame(raw, "crc16").ok).toBe(true);
  });

  it("a frame that announces more bytes than were supplied", () => {
    const raw = [0x68, 0x20, 0x68, 0x00, 0x01, 0x01, 0x02, 0x16];

    expect(decodeFrame(raw, "sum").issues.map((issue) => issue.code)).toContain("tooShort");
  });

  it("a corrupted checksum", () => {
    const raw = [...READ_TIME_REQUEST];
    raw[raw.length - 2] = raw[raw.length - 2]! ^ 0xff;
    const decoded = decodeFrame(raw, "sum");

    expect(decoded.ok).toBe(false);
    const issue = decoded.issues.find((item) => item.code === "invalidChecksum");
    expect(issue?.expected).toEqual([0x02]);
    expect(issue?.actual).toEqual([0x02 ^ 0xff]);
  });

  it("a bad end byte", () => {
    const raw = [...READ_TIME_REQUEST];
    raw[raw.length - 1] = 0x00;
    const decoded = decodeFrame(raw, "sum");

    expect(decoded.issues.map((issue) => issue.code)).toContain("invalidEnd");
  });

  it("a sum frame read as crc16", () => {
    expect(decodeFrame(READ_TIME_REQUEST, "crc16").ok).toBe(false);
  });

  it("still maps the fields it could resolve", () => {
    const raw = [...READ_TIME_REQUEST];
    raw[raw.length - 1] = 0x00;

    expect(decodeFrame(raw, "sum").fields.length).toBeGreaterThan(0);
  });

  it("exposes the minimum frame length it enforces", () => {
    expect(MIN_FRAME_LENGTH).toBe(7);
  });
});

describe("scanFrames", () => {
  const frame = encodeFrame(0x00, 0x01, [0x01], "sum");

  it("extracts a single frame", () => {
    const result = scanFrames(frame, "sum");

    expect(result.frames).toHaveLength(1);
    expect(result.frames[0]?.ok).toBe(true);
    expect(result.rest).toEqual([]);
  });

  it("extracts several frames from one chunk", () => {
    const result = scanFrames([...frame, ...frame, ...frame], "sum");

    expect(result.frames.filter((item) => item.ok)).toHaveLength(3);
    expect(result.rest).toEqual([]);
  });

  it("discards noise before the start byte", () => {
    const result = scanFrames([0xaa, 0xbb, 0xcc, ...frame], "sum");

    expect(result.frames.filter((item) => item.ok)).toHaveLength(1);
  });

  it("keeps a partial tail for the next chunk", () => {
    const partial = frame.slice(0, 4);
    const result = scanFrames([...frame, ...partial], "sum");

    expect(result.frames.filter((item) => item.ok)).toHaveLength(1);
    expect(result.rest).toEqual(partial);
  });

  it("resynchronises past a corrupted frame", () => {
    const corrupt = [...frame];
    corrupt[corrupt.length - 2] = corrupt[corrupt.length - 2]! ^ 0xff;
    const result = scanFrames([...corrupt, ...frame], "sum");

    expect(result.frames.some((item) => !item.ok)).toBe(true);
    expect(result.frames.some((item) => item.ok)).toBe(true);
  });

  it("returns nothing for a buffer with no start byte", () => {
    const result = scanFrames([0xaa, 0xbb], "sum");

    expect(result.frames).toEqual([]);
    expect(result.rest).toEqual([]);
  });

  it("returns nothing for an empty buffer", () => {
    expect(scanFrames([], "sum")).toEqual({ frames: [], rest: [] });
  });

  it("stops rather than looping on unusable noise", () => {
    const result = scanFrames([START_BYTE, START_BYTE, START_BYTE], "sum");

    expect(result.frames).toEqual([]);
    expect(result.rest.length).toBeLessThanOrEqual(3);
  });
});
