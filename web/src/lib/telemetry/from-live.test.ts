import { gatewayStatusSchema, historySchema, fleetSchema, eventSchema } from "@/lib/api/schema";
import {
  eventFixture,
  fleetFixture,
  gatewayStatusFixture,
  historyFixture,
} from "@/test/utils";

import {
  classifyError,
  instantToMs,
  toChecksumMode,
  toClockReport,
  toClockState,
  toCounters,
  toFleet,
  toHealthState,
  toHealthView,
  toIdentity,
  toLogEvent,
  toLogEvents,
  toSchedule,
  toScheduleMode,
  toSkewSamples,
  toTicks,
} from "./from-live";

/**
 * Every fixture goes through the schema first.
 *
 * These adapters take parsed payloads, and parsing is where defaults are
 * applied. Feeding them a hand-written object would test the adapters against
 * a shape the API never produces.
 */
const status = gatewayStatusSchema.parse(gatewayStatusFixture());
const history = historySchema.parse(historyFixture());
const fleet = fleetSchema.parse(fleetFixture());

const parseEvent = (overrides: Record<string, unknown> = {}) =>
  eventSchema.parse(eventFixture(overrides));

describe("instants", () => {
  it("reads an RFC3339 instant", () => {
    expect(instantToMs("2026-09-02T12:00:00Z")).toBe(Date.UTC(2026, 8, 2, 12, 0, 0));
  });

  it("reads an absent instant as never, not as the epoch", () => {
    // Zero is what every readout tests for; NaN would propagate into a
    // duration and print as a dash in one place and "NaN" in another.
    expect(instantToMs(undefined)).toBe(0);
    expect(instantToMs("")).toBe(0);
    expect(instantToMs("not a date")).toBe(0);
  });
});

describe("state names", () => {
  it.each(["unknown", "online", "degraded", "offline"] as const)("keeps %s", (state) => {
    expect(toHealthState(state)).toBe(state);
  });

  it.each(["unknown", "ok", "warn", "critical"] as const)("keeps the clock state %s", (state) => {
    expect(toClockState(state)).toBe(state);
  });

  it("falls back to unknown for a state this console cannot colour", () => {
    // A newer gateway may invent a state; "unknown" is exactly what the
    // console knows about it, and it has a colour and a word in both locales.
    expect(toHealthState("quarantined")).toBe("unknown");
    expect(toClockState("skewed-ish")).toBe("unknown");
  });

  it("falls back to sum for an unrecognised checksum mode", () => {
    expect(toChecksumMode("crc16")).toBe("crc16");
    expect(toChecksumMode("sum")).toBe("sum");
    expect(toChecksumMode("unspecified")).toBe("sum");
  });

  it("treats anything but aligned as an interval schedule", () => {
    expect(toScheduleMode("aligned")).toBe("aligned");
    expect(toScheduleMode("interval")).toBe("interval");
    expect(toScheduleMode("")).toBe("interval");
  });
});

describe("error wording", () => {
  it.each([
    ["decode frame: invalid checksum: sum mismatch", "invalidChecksum"],
    ["no response frame received", "noResponse"],
    ["read tcp 127.0.0.1:9000: i/o timeout", "timeout"],
    ["read: EOF", "connectionClosed"],
    ["connection reset by peer", "connectionClosed"],
  ] as const)("recognises %s", (message, code) => {
    expect(classifyError(message)).toBe(code);
  });

  it("leaves a fault it has no word for untranslated", () => {
    // Better an untranslated message than one filed under the wrong fault:
    // the reader can match the gateway's own words against its log.
    expect(classifyError("adapter address rejected by device")).toBeUndefined();
  });
});

describe("log events", () => {
  it("decodes the hex the gateway recorded", () => {
    const event = toLogEvent(parseEvent());

    expect(event.bytes).toEqual([0x68, 0x03, 0x68, 0x00, 0x01, 0x01, 0x02, 0x16]);
    expect(event.command).toBe("read-time");
    expect(event.at).toBe(Date.UTC(2026, 8, 2, 12, 0, 0));
  });

  it.each([
    ["TX", "tx", "gateway"],
    ["RX", "rx", "device"],
    ["ERR", "err", "gateway"],
    ["SYSTEM", "sys", "gateway"],
  ] as const)("puts a %s row in the right lane", (direction, expected, lane) => {
    const event = toLogEvent(parseEvent({ direction }));

    expect(event.direction).toBe(expected);
    // The gateway writes down both halves of the exchange, so the lane comes
    // from the direction rather than from who recorded it.
    expect(event.source).toBe(lane);
  });

  it("files an unrecognised direction as a system note", () => {
    expect(toLogEvent(parseEvent({ direction: "SIDEWAYS" })).direction).toBe("sys");
  });

  it("translates an error it recognises and keeps the original words", () => {
    const event = toLogEvent(
      parseEvent({ direction: "ERR", rawHex: "", error: "decode: invalid checksum" }),
    );

    expect(event.errorCode).toBe("invalidChecksum");
    expect(event.detail).toBe("decode: invalid checksum");
  });

  it("carries a system message as free text", () => {
    const event = toLogEvent(
      parseEvent({ direction: "SYSTEM", rawHex: "", command: "device-state", message: "online -> degraded" }),
    );

    expect(event.errorCode).toBeUndefined();
    expect(event.detail).toBe("online -> degraded");
  });

  it("reads zero cycle and attempt rather than guessing them", () => {
    const event = toLogEvent(parseEvent());

    // The gateway's history records neither; inferring them from neighbouring
    // rows would put a made-up retry number in front of an operator.
    expect(event.cycle).toBe(0);
    expect(event.attempt).toBe(0);
    expect(event.latencyMs).toBeUndefined();
  });

  it("orders the log oldest first", () => {
    const events = toLogEvents([
      eventFixture({ id: 2, timestamp: "2026-09-02T12:00:02Z" }),
      eventFixture({ id: 1, timestamp: "2026-09-02T12:00:01Z" }),
      eventFixture({ id: 3, timestamp: "2026-09-02T12:00:03Z" }),
    ].map((event) => eventSchema.parse(event)));

    expect(events.map((event) => event.id)).toEqual([1, 2, 3]);
  });

  it("breaks a tie on the same instant by id", () => {
    const events = toLogEvents(
      [eventFixture({ id: 9 }), eventFixture({ id: 4 })].map((event) => eventSchema.parse(event)),
    );

    expect(events.map((event) => event.id)).toEqual([4, 9]);
  });

  it("leaves a row with no frame empty rather than decoding nothing", () => {
    expect(toLogEvent(parseEvent({ rawHex: "" })).bytes).toEqual([]);
  });
});

describe("schedule ticks", () => {
  it("takes every read-time request, retries included", () => {
    const events = toLogEvents(
      [
        eventFixture({ id: 1, direction: "TX", command: "read-time", timestamp: "2026-09-02T12:00:00Z" }),
        eventFixture({ id: 2, direction: "RX", command: "read-time", timestamp: "2026-09-02T12:00:01Z" }),
        eventFixture({ id: 3, direction: "TX", command: "read-identity", timestamp: "2026-09-02T12:00:02Z" }),
        eventFixture({ id: 4, direction: "TX", command: "read-time", timestamp: "2026-09-02T12:00:03Z" }),
      ].map((event) => eventSchema.parse(event)),
    );

    // A retried poll shows as a cluster. That is a truthful rendering of what
    // went out on the line -- the gateway does not record which attempt a
    // frame was, and inventing one would be worse than a dense tick.
    expect(toTicks(events)).toEqual([
      Date.UTC(2026, 8, 2, 12, 0, 0),
      Date.UTC(2026, 8, 2, 12, 0, 3),
    ]);
  });
});

describe("readouts", () => {
  it("takes the clock report as the gateway computed it", () => {
    const clock = toClockReport(status);

    // Not recomputed from the history: the gateway's window is longer than
    // what it sends, and a recomputed median would quietly disagree with the
    // figure the gateway made its own decision on.
    expect(clock).toEqual({
      state: "warn",
      skewMs: -2500,
      medianMs: -2400,
      minMs: -3000,
      maxMs: -1800,
      driftPerDayMs: 24_000,
      driftDetermined: true,
      fit: 0.94,
      samples: 31,
      roundTripMs: 9,
    });
  });

  it("keeps the sample window in order and signed", () => {
    expect(toSkewSamples(history)).toEqual([
      { at: Date.UTC(2026, 8, 2, 11, 59, 0), skewMs: -2600, roundTripMs: 8 },
      { at: Date.UTC(2026, 8, 2, 12, 0, 0), skewMs: -2500, roundTripMs: 9 },
    ]);
  });

  it("builds the outcome strip from the history and the rest from the status", () => {
    const health = toHealthView(status, history);

    expect(health.state).toBe("degraded");
    expect(health.window).toEqual([true, false]);
    expect(health.availability).toBe(0.875);
    // The window sent is a truncated copy; the sample count is the gateway's.
    expect(health.windowSamples).toBe(40);
    expect(health.latencyP50Ms).toBe(44);
    expect(health.latencyP95Ms).toBe(45);
  });

  it("maps the counters, using connection attempts for sessions", () => {
    expect(toCounters(status)).toEqual({
      successfulReads: 42,
      failedReads: 3,
      retries: 7,
      protocolErrors: 2,
      reconnects: 1,
      connections: 1,
    });
  });

  it("reads the schedule", () => {
    expect(toSchedule(status)).toEqual({
      mode: "aligned",
      intervalMs: 60_000,
      offsetMs: 5000,
      nextPollAt: Date.UTC(2026, 8, 2, 12, 1, 5),
    });
  });

  it("falls back to the flat interval when the schedule section is empty", () => {
    const older = gatewayStatusSchema.parse(
      gatewayStatusFixture({ schedule: { mode: "", intervalMs: 0, offsetMs: 0 } }),
    );

    expect(toSchedule(older).intervalMs).toBe(5000);
  });

  it("reports no identity until the probe has read one", () => {
    const unread = gatewayStatusSchema.parse(
      gatewayStatusFixture({ identity: { known: false, supported: true } }),
    );

    expect(toIdentity(unread)).toBeNull();
    expect(toIdentity(status)).toEqual({ model: "TTR20", serial: "SN-42", firmware: "1.2.3" });
  });
});

describe("fleet", () => {
  it("lists every device with its own states", () => {
    const view = toFleet(fleet);

    expect(view.devices.map((device) => device.id)).toEqual(["alpha", "beta"]);
    expect(view.devices[0]!.health).toBe("degraded");
    expect(view.devices[1]!.clock).toBe("ok");
    expect(view.devices[1]!.skewMs).toBe(12);
  });

  it("names a device by its id when it has no name", () => {
    const unnamed = fleetSchema.parse({
      devices: [gatewayStatusFixture({ deviceId: "solo", deviceName: "" })],
    });

    expect(toFleet(unnamed).devices[0]!.name).toBe("solo");
  });

  it.each([
    ["running", true],
    ["degraded", true],
    ["stopped", false],
  ] as const)("treats a %s device as polling=%s", (state, polling) => {
    // Degraded means it is still trying, which is what the dot is for.
    const one = fleetSchema.parse({ devices: [gatewayStatusFixture({ state })] });

    expect(toFleet(one).devices[0]!.running).toBe(polling);
  });

  it("carries the summary, including which clock is furthest out", () => {
    const view = toFleet(fleet);

    expect(view.online).toBe(1);
    expect(view.degraded).toBe(1);
    expect(view.clockWarn).toBe(1);
    expect(view.worstDeviceId).toBe("alpha");
    expect(view.worstSkewMs).toBe(-2500);
  });

  it("handles a gateway that reports no devices", () => {
    const view = toFleet(fleetSchema.parse({}));

    expect(view.devices).toEqual([]);
    expect(view.worstDeviceId).toBe("");
  });
});
