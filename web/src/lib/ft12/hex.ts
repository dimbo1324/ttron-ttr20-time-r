/**
 * Hex <-> bytes, written for a field that a human is typing into.
 *
 * The parser is deliberately forgiving about separators (spaces, commas,
 * newlines, `0x` prefixes) because operators paste captures from every tool
 * that ever printed a frame, and strict about digits, because a typo there is
 * a different frame rather than a formatting preference.
 */

export interface HexParseResult {
  bytes: number[];
  /** Offsets in the *cleaned* string that were not valid hex digits. */
  invalid: string[];
  /** True when the cleaned input has an odd number of digits. */
  truncated: boolean;
}

const HEX_DIGITS = /^[0-9a-fA-F]+$/;

export function parseHex(input: string): HexParseResult {
  const tokens = input
    .replace(/0[xX]/g, " ")
    .split(/[\s,;:_-]+/)
    .filter((token) => token.length > 0);

  const invalid: string[] = [];
  let digits = "";
  for (const token of tokens) {
    if (HEX_DIGITS.test(token)) digits += token;
    else invalid.push(token);
  }

  const bytes: number[] = [];
  // A trailing half-byte is reported through `truncated` rather than guessed at.
  for (let index = 0; index + 1 < digits.length; index += 2) {
    bytes.push(Number.parseInt(digits.slice(index, index + 2), 16));
  }

  return { bytes, invalid, truncated: digits.length % 2 === 1 };
}

export function toHex(byte: number, width = 2): string {
  return byte.toString(16).toUpperCase().padStart(width, "0");
}

export function formatHex(bytes: readonly number[]): string {
  return bytes.map((byte) => toHex(byte)).join(" ");
}

export function formatBinary(byte: number): string {
  return byte.toString(2).padStart(8, "0");
}

/** Printable ASCII for the byte grid; everything else shows as a middle dot. */
export function asciiGlyph(byte: number): string {
  return byte >= 0x20 && byte <= 0x7e ? String.fromCharCode(byte) : "·";
}
