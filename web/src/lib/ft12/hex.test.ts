import { asciiGlyph, formatBinary, formatHex, parseHex, toHex } from "./hex";

describe("parseHex", () => {
  it.each([
    { name: "spaces", input: "68 03 68" },
    { name: "commas", input: "68,03,68" },
    { name: "newlines", input: "68\n03\n68" },
    { name: "colons", input: "68:03:68" },
    { name: "dashes", input: "68-03-68" },
    { name: "0x prefixes", input: "0x68 0x03 0x68" },
    { name: "no separators", input: "680368" },
  ])("accepts $name", ({ input }) => {
    expect(parseHex(input).bytes).toEqual([0x68, 0x03, 0x68]);
  });

  it("is case-insensitive", () => {
    expect(parseHex("ab CD ef").bytes).toEqual([0xab, 0xcd, 0xef]);
  });

  it("reports tokens that are not hex", () => {
    const result = parseHex("68 zz 03");
    expect(result.invalid).toEqual(["zz"]);
    expect(result.bytes).toEqual([0x68, 0x03]);
  });

  it("reports an odd number of digits", () => {
    const result = parseHex("68 0");
    expect(result.truncated).toBe(true);
    expect(result.bytes).toEqual([0x68]);
  });

  it("returns nothing for empty input", () => {
    expect(parseHex("")).toEqual({ bytes: [], invalid: [], truncated: false });
    expect(parseHex("   \n  ").bytes).toEqual([]);
  });
});

describe("toHex", () => {
  it.each([
    { value: 0x00, want: "00" },
    { value: 0x0f, want: "0F" },
    { value: 0xff, want: "FF" },
    { value: 0x68, want: "68" },
  ])("renders $value as $want", ({ value, want }) => {
    expect(toHex(value)).toBe(want);
  });

  it("honours an explicit width", () => {
    expect(toHex(0x1, 4)).toBe("0001");
  });
});

describe("formatHex", () => {
  it("joins bytes with single spaces", () => {
    expect(formatHex([0x68, 0x03, 0x16])).toBe("68 03 16");
  });

  it("returns an empty string for no bytes", () => {
    expect(formatHex([])).toBe("");
  });

  it("round-trips through parseHex", () => {
    const bytes = [0x68, 0x03, 0x68, 0x00, 0x01, 0x01, 0x02, 0x16];
    expect(parseHex(formatHex(bytes)).bytes).toEqual(bytes);
  });
});

describe("formatBinary", () => {
  it.each([
    { value: 0x00, want: "00000000" },
    { value: 0x01, want: "00000001" },
    { value: 0x80, want: "10000000" },
    { value: 0xff, want: "11111111" },
  ])("renders $value", ({ value, want }) => {
    expect(formatBinary(value)).toBe(want);
  });
});

describe("asciiGlyph", () => {
  it.each([
    { value: 0x41, want: "A" },
    { value: 0x20, want: " " },
    { value: 0x7e, want: "~" },
  ])("renders printable $value", ({ value, want }) => {
    expect(asciiGlyph(value)).toBe(want);
  });

  it.each([0x00, 0x1f, 0x7f, 0xff])("replaces unprintable %s", (value) => {
    expect(asciiGlyph(value)).toBe("·");
  });
});
