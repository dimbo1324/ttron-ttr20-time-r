import {
  checksumLength,
  checksumMatches,
  computeChecksum,
  type ChecksumMode,
} from "./checksum";

/**
 * FT1.2-like variable-length frame, mirroring `internal/protocol/frame`.
 *
 *   0x68 | LEN | 0x68 | CONTROL | ADDRESS | DATA... | CHECKSUM | 0x16
 *
 * `LEN` counts CONTROL + ADDRESS + DATA. The checksum covers the same span,
 * which is why decoding is mode-aware: a `crc16` frame is one byte longer than
 * a `sum` frame carrying identical data.
 */

export const START_BYTE = 0x68;
export const END_BYTE = 0x16;

/** Smallest legal frame: header(3) + control + address + checksum(1) + end. */
export const MIN_FRAME_LENGTH = 7;

/**
 * Largest frame the format can express.
 *
 * LEN is a single byte, so a payload tops out at 255: header(3) + 255 +
 * crc16(2) + end(1). There is deliberately no "frame too large" branch in the
 * decoder — the bound is structural rather than a limit worth enforcing, and a
 * guard that cannot fire is a guard nobody can test.
 */
export const MAX_FRAME_LENGTH = 3 + 0xff + 2 + 1;

export interface Ft12Frame {
  control: number;
  address: number;
  data: number[];
  raw: number[];
}

/**
 * What each byte of a frame *is*.
 *
 * The inspector renders this directly, so the decoder produces it even for
 * frames that fail validation — a reader learning the format needs to see
 * which byte was wrong, not a bare "invalid frame".
 */
export type FieldKind =
  | "start"
  | "length"
  | "startRepeat"
  | "control"
  | "address"
  | "data"
  | "checksum"
  | "end"
  | "trailing";

export interface FrameField {
  kind: FieldKind;
  offset: number;
  bytes: number[];
}

export type FrameErrorCode =
  | "tooShort"
  | "invalidStart"
  | "invalidStartRepeat"
  | "invalidLength"
  | "invalidChecksum"
  | "invalidEnd";

export interface FrameIssue {
  code: FrameErrorCode;
  /** Byte the problem points at, when there is one to point at. */
  offset?: number;
  expected?: number[];
  actual?: number[];
}

export interface DecodedFrame {
  ok: boolean;
  mode: ChecksumMode;
  fields: FrameField[];
  issues: FrameIssue[];
  frame?: Ft12Frame;
  /** Checksum the payload should have carried — shown next to the actual one. */
  expectedChecksum: number[];
  consumed: number;
}

export function payloadBytes(control: number, address: number, data: readonly number[]): number[] {
  return [control & 0xff, address & 0xff, ...data.map((byte) => byte & 0xff)];
}

export function encodeFrame(
  control: number,
  address: number,
  data: readonly number[],
  mode: ChecksumMode,
): number[] {
  const payload = payloadBytes(control, address, data);
  return [
    START_BYTE,
    payload.length & 0xff,
    START_BYTE,
    ...payload,
    ...computeChecksum(mode, payload),
    END_BYTE,
  ];
}

/**
 * Decodes one frame from the start of `raw`.
 *
 * Never throws: every failure mode comes back as an issue plus whatever field
 * map could still be derived, because the inspector's job is to explain a bad
 * frame rather than refuse it.
 */
export function decodeFrame(raw: readonly number[], mode: ChecksumMode): DecodedFrame {
  const sumLength = checksumLength(mode);
  const issues: FrameIssue[] = [];
  const fields: FrameField[] = [];
  const expectedChecksum: number[] = [];

  if (raw.length < MIN_FRAME_LENGTH) {
    return {
      ok: false,
      mode,
      fields,
      issues: [{ code: "tooShort", actual: [...raw] }],
      expectedChecksum,
      consumed: 0,
    };
  }

  if (raw[0] !== START_BYTE) {
    issues.push({ code: "invalidStart", offset: 0, expected: [START_BYTE], actual: [raw[0]!] });
  }
  fields.push({ kind: "start", offset: 0, bytes: [raw[0]!] });

  const payloadLength = raw[1]!;
  fields.push({ kind: "length", offset: 1, bytes: [payloadLength] });

  if (raw[2] !== START_BYTE) {
    issues.push({
      code: "invalidStartRepeat",
      offset: 2,
      expected: [START_BYTE],
      actual: [raw[2]!],
    });
  }
  fields.push({ kind: "startRepeat", offset: 2, bytes: [raw[2]!] });

  // CONTROL and ADDRESS are always present, so anything below 2 cannot describe
  // a real payload and there is no safe way to lay out the rest of the frame.
  if (payloadLength < 2) {
    issues.push({ code: "invalidLength", offset: 1, actual: [payloadLength] });
    return { ok: false, mode, fields, issues, expectedChecksum, consumed: 0 };
  }

  const total = 3 + payloadLength + sumLength + 1;
  if (raw.length < total) {
    issues.push({ code: "tooShort", actual: [...raw] });
    return { ok: false, mode, fields, issues, expectedChecksum, consumed: 0 };
  }

  const control = raw[3]!;
  const address = raw[4]!;
  const data = raw.slice(5, 3 + payloadLength);

  fields.push({ kind: "control", offset: 3, bytes: [control] });
  fields.push({ kind: "address", offset: 4, bytes: [address] });
  if (data.length > 0) {
    fields.push({ kind: "data", offset: 5, bytes: [...data] });
  }

  const checksumOffset = 3 + payloadLength;
  const checksum = raw.slice(checksumOffset, checksumOffset + sumLength);
  const payload = raw.slice(3, 3 + payloadLength);
  expectedChecksum.push(...computeChecksum(mode, payload));

  fields.push({ kind: "checksum", offset: checksumOffset, bytes: [...checksum] });
  if (!checksumMatches(mode, payload, checksum)) {
    issues.push({
      code: "invalidChecksum",
      offset: checksumOffset,
      expected: expectedChecksum,
      actual: [...checksum],
    });
  }

  const endOffset = checksumOffset + sumLength;
  const end = raw[endOffset]!;
  fields.push({ kind: "end", offset: endOffset, bytes: [end] });
  if (end !== END_BYTE) {
    issues.push({ code: "invalidEnd", offset: endOffset, expected: [END_BYTE], actual: [end] });
  }

  if (raw.length > total) {
    fields.push({ kind: "trailing", offset: total, bytes: raw.slice(total) });
  }

  const ok = issues.length === 0;
  return {
    ok,
    mode,
    fields,
    issues,
    frame: ok ? { control, address, data: [...data], raw: raw.slice(0, total) } : undefined,
    expectedChecksum,
    consumed: total,
  };
}

/**
 * Pulls every complete frame out of a byte stream, the way the Go
 * `StreamParser` does: noise before a start byte is discarded, a partial tail
 * is kept for the next chunk, and a frame that fails to decode costs one byte
 * of resynchronisation rather than the whole buffer.
 */
export function scanFrames(
  raw: readonly number[],
  mode: ChecksumMode,
): { frames: DecodedFrame[]; rest: number[] } {
  const frames: DecodedFrame[] = [];
  let buffer = [...raw];

  while (buffer.length > 0) {
    const start = buffer.indexOf(START_BYTE);
    if (start < 0) return { frames, rest: [] };
    if (start > 0) buffer = buffer.slice(start);
    if (buffer.length < MIN_FRAME_LENGTH) break;

    const decoded = decodeFrame(buffer, mode);
    if (decoded.ok && decoded.consumed > 0) {
      frames.push(decoded);
      buffer = buffer.slice(decoded.consumed);
      continue;
    }
    if (decoded.issues.some((issue) => issue.code === "tooShort")) break;

    frames.push(decoded);
    buffer = buffer.slice(1);
  }

  return { frames, rest: buffer };
}
