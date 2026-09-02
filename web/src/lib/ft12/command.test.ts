import {
  buildReadIdentityRequest,
  buildReadIdentityResponse,
  buildReadTimeRequest,
  buildReadTimeResponse,
  commandById,
  commandName,
  COMMANDS,
  formatDeviceTime,
  IDENTITY_SEPARATOR,
  parsePayload,
  READ_IDENTITY,
  READ_TIME,
  READ_TIME_LAYOUT,
  RESPONSE_BIT,
} from "./command";

const ascii = (text: string) => [...text].map((char) => char.charCodeAt(0));

describe("command registry", () => {
  it("knows the two first-class commands", () => {
    expect(COMMANDS.map((command) => command.id)).toEqual([READ_TIME, READ_IDENTITY]);
  });

  it.each([
    { id: READ_TIME, name: "read-time" },
    { id: READ_IDENTITY, name: "read-identity" },
  ])("resolves 0x0$id to $name", ({ id, name }) => {
    expect(commandById(id)?.name).toBe(name);
    expect(commandName(id)).toBe(name);
  });

  it("renders an unknown id rather than failing", () => {
    expect(commandById(0x7f)).toBeUndefined();
    expect(commandName(0x7f)).toBe("unknown-0x7F");
    expect(commandName(0x00)).toBe("unknown-0x00");
  });

  it("gives every command a description key the dictionary can resolve", () => {
    for (const command of COMMANDS) {
      expect(["readTime", "readIdentity"]).toContain(command.descriptionKey);
    }
  });
});

describe("read-time payloads", () => {
  it("builds a one-byte request", () => {
    expect(buildReadTimeRequest()).toEqual([READ_TIME]);
  });

  it("builds a response carrying the timestamp as ASCII", () => {
    const payload = buildReadTimeResponse(new Date(2026, 5, 2, 12, 34, 56));

    expect(payload[0]).toBe(READ_TIME);
    expect(String.fromCharCode(...payload.slice(1))).toBe("2026-06-02 12:34:56");
    expect(payload).toHaveLength(1 + READ_TIME_LAYOUT.length);
  });

  it("parses a request", () => {
    expect(parsePayload(buildReadTimeRequest(), false)).toMatchObject({
      kind: "readTimeRequest",
      commandName: "read-time",
    });
  });

  it("parses a response", () => {
    const payload = buildReadTimeResponse(new Date(2026, 5, 2, 12, 34, 56));
    const parsed = parsePayload(payload, true);

    expect(parsed.kind).toBe("readTimeResponse");
    expect(parsed.value).toBe("2026-06-02 12:34:56");
    expect(parsed.error).toBeUndefined();
  });

  it("flags a timestamp of the wrong length", () => {
    const parsed = parsePayload([READ_TIME, ...ascii("2026-06-02")], true);

    expect(parsed.error).toBe("invalidLengthPayload");
  });

  it("flags a body of the right length that is not a timestamp", () => {
    const parsed = parsePayload([READ_TIME, ...ascii("not-a-timestamp!!!!")], true);

    expect(parsed.value).toBe("not-a-timestamp!!!!");
    expect(parsed.error).toBe("invalidTimestamp");
  });
});

describe("read-identity payloads", () => {
  it("builds a one-byte request", () => {
    expect(buildReadIdentityRequest()).toEqual([READ_IDENTITY]);
  });

  it("joins the identity triple with the separator", () => {
    const payload = buildReadIdentityResponse("TTR20", "SN-1", "1.0");

    expect(payload[0]).toBe(READ_IDENTITY);
    expect(String.fromCharCode(...payload.slice(1))).toBe(
      ["TTR20", "SN-1", "1.0"].join(IDENTITY_SEPARATOR),
    );
  });

  it("parses a request", () => {
    expect(parsePayload(buildReadIdentityRequest(), false).kind).toBe("readIdentityRequest");
  });

  it("splits a response into labelled fields", () => {
    const parsed = parsePayload(buildReadIdentityResponse("TTR20", "SN-42", "1.2.3"), true);

    expect(parsed.kind).toBe("readIdentityResponse");
    expect(parsed.error).toBeUndefined();
    expect(parsed.fields).toEqual([
      { label: "model", value: "TTR20" },
      { label: "serial", value: "SN-42" },
      { label: "firmware", value: "1.2.3" },
    ]);
  });

  it.each([
    { name: "too few fields", body: "TTR20|SN-1" },
    { name: "too many fields", body: "a|b|c|d" },
    { name: "a blank field", body: "TTR20|   |1.0" },
    { name: "an empty leading field", body: "|SN-1|1.0" },
  ])("rejects $name", ({ body }) => {
    expect(parsePayload([READ_IDENTITY, ...ascii(body)], true).error).toBe("invalidIdentity");
  });
});

describe("parsePayload fallbacks", () => {
  it("reports an empty payload", () => {
    const parsed = parsePayload([], false);

    expect(parsed.kind).toBe("unknown");
    expect(parsed.error).toBe("emptyPayload");
    expect(parsed.commandName).toBe("—");
  });

  it("recognises an ACK response", () => {
    const parsed = parsePayload([0x33, ...ascii("OK")], true);

    expect(parsed.kind).toBe("ack");
    expect(parsed.value).toBe("OK");
    expect(parsed.commandName).toBe("unknown-0x33");
  });

  it("returns unknown for an unrecognised command", () => {
    const parsed = parsePayload([0x33, ...ascii("data")], false);

    expect(parsed.kind).toBe("unknown");
    expect(parsed.value).toBe("data");
  });

  it("leaves the value undefined when an unknown command has no body", () => {
    expect(parsePayload([0x33], false).value).toBeUndefined();
  });

  it("separates a request from a response by direction alone", () => {
    const payload = buildReadTimeRequest();

    expect(parsePayload(payload, false).kind).toBe("readTimeRequest");
    expect(parsePayload(payload, true).kind).toBe("readTimeResponse");
  });
});

describe("formatDeviceTime", () => {
  it("renders the wire layout", () => {
    expect(formatDeviceTime(new Date(2026, 0, 2, 3, 4, 5))).toBe("2026-01-02 03:04:05");
  });

  it("pads every component", () => {
    expect(formatDeviceTime(new Date(2026, 8, 9, 9, 9, 9))).toBe("2026-09-09 09:09:09");
  });

  it("round-trips through the response payload", () => {
    const date = new Date(2026, 11, 31, 23, 59, 58);
    const parsed = parsePayload(buildReadTimeResponse(date), true);

    expect(parsed.value).toBe(formatDeviceTime(date));
  });
});

describe("RESPONSE_BIT", () => {
  it("is the documented 0x80", () => {
    expect(RESPONSE_BIT).toBe(0x80);
  });
});
