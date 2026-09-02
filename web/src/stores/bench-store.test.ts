import { EMPTY_HEALTH } from "@/lib/bench/domain";
import { decodeFrame, formatHex, READ_IDENTITY, READ_TIME, RESPONSE_BIT } from "@/lib/ft12";

import {
  DEFAULT_FAULTS,
  DEFAULT_GATEWAY,
  selectActiveFaultCount,
  selectAvailability,
  selectLastFrameHex,
  useBenchStore,
  type BenchEvent,
} from "./bench-store";

/**
 * The engine is driven by wall time and by a random draw for line jitter and
 * fault probability. Both are pinned here: a simulation that cannot be
 * replayed is a simulation whose failures cannot be read.
 *
 * `Math.random() === 0` makes every probability comparison (`random < p`)
 * true for any non-zero probability and false for zero, which is exactly the
 * on/off control a fault test wants, and it fixes jitter at its floor.
 */
const T0 = Date.UTC(2026, 8, 2, 12, 0, 0);
let now = T0;

function resetStore() {
  useBenchStore.setState({
    running: false,
    connected: false,
    checksumMode: "sum",
    adapterAddress: 1,
    faults: { ...DEFAULT_FAULTS },
    identity: { model: "TTR20", serial: "SN-0000042", firmware: "1.2.3" },
    gateway: { ...DEFAULT_GATEWAY, identityProbe: false },
    events: [],
    samples: [],
    health: EMPTY_HEALTH,
    counters: {
      successfulReads: 0,
      failedReads: 0,
      reconnects: 0,
      retries: 0,
      protocolErrors: 0,
      exhaustedPolls: 0,
      connections: 0,
    },
    cycle: 0,
    nextPollAt: 0,
    lastCycleAt: 0,
    lastDeviceTime: null,
    identityRead: null,
    clockOrigin: T0,
  });
}

const state = () => useBenchStore.getState();
const events = () => state().events;
const byDirection = (direction: BenchEvent["direction"]) =>
  events().filter((event) => event.direction === direction);
const commands = () => events().map((event) => `${event.direction}:${event.command}`);

/** Runs one polling cycle from a known instant. */
function runOneCycle() {
  state().tick(state().nextPollAt);
}

beforeEach(() => {
  now = T0;
  jest.spyOn(Date, "now").mockImplementation(() => now);
  jest.spyOn(Math, "random").mockReturnValue(0);
  resetStore();
});

describe("start and stop", () => {
  it("polls immediately in interval mode", () => {
    useBenchStore.getState().patchGateway({ scheduleMode: "interval", intervalMs: 5000 });
    state().start();

    expect(state().running).toBe(true);
    expect(state().connected).toBe(true);
    expect(state().nextPollAt).toBe(T0);
    expect(state().counters.connections).toBe(1);
    expect(commands()).toContain("sys:connect");
  });

  it("waits for the boundary in aligned mode", () => {
    useBenchStore
      .getState()
      .patchGateway({ scheduleMode: "aligned", intervalMs: 60_000, offsetMs: 5000 });
    state().start();

    expect(state().nextPollAt).toBeGreaterThan(T0);
    expect(new Date(state().nextPollAt).getUTCSeconds()).toBe(5);
  });

  it("ignores a second start", () => {
    state().start();
    const first = state().nextPollAt;
    state().start();

    expect(state().counters.connections).toBe(1);
    expect(state().nextPollAt).toBe(first);
  });

  it("stops and forgets the identity read on this connection", () => {
    state().start();
    runOneCycle();
    useBenchStore.setState({ identityRead: { model: "m", serial: "s", firmware: "f" } });

    state().stop();

    expect(state().running).toBe(false);
    expect(state().connected).toBe(false);
    expect(state().identityRead).toBeNull();
    expect(commands()).toContain("sys:disconnect");
  });

  it("ignores a stop when not running", () => {
    state().stop();

    expect(events()).toEqual([]);
  });
});

describe("a healthy polling cycle", () => {
  beforeEach(() => {
    useBenchStore.getState().patchGateway({ scheduleMode: "interval", intervalMs: 5000 });
    state().start();
    runOneCycle();
  });

  it("emits a request, a response and the state changes they caused", () => {
    expect(commands()).toEqual([
      "sys:connect",
      "tx:read-time",
      "rx:read-time",
      // The first successful cycle moves both state machines off "unknown".
      "sys:device-state",
      "sys:clock-state",
    ]);
  });

  it("sends a decodable read-time request", () => {
    const [request] = byDirection("tx");
    const decoded = decodeFrame(request!.bytes, "sum");

    expect(decoded.ok).toBe(true);
    expect(decoded.frame?.control).toBe(0x00);
    expect(decoded.frame?.data).toEqual([READ_TIME]);
  });

  it("answers with the response bit set", () => {
    const [response] = byDirection("rx");
    const decoded = decodeFrame(response!.bytes, "sum");

    expect(decoded.ok).toBe(true);
    expect(decoded.frame!.control & RESPONSE_BIT).toBe(RESPONSE_BIT);
  });

  it("counts the read and records a skew sample", () => {
    expect(state().counters.successfulReads).toBe(1);
    expect(state().counters.failedReads).toBe(0);
    expect(state().samples).toHaveLength(1);
    expect(state().lastDeviceTime).not.toBeNull();
  });

  it("marks the device online", () => {
    expect(state().health.state).toBe("online");
    expect(selectAvailability(state())).toBe(1);
  });

  it("schedules the next poll", () => {
    expect(state().nextPollAt).toBeGreaterThan(T0);
  });

  it("does nothing when ticked before the next poll", () => {
    const before = events().length;
    state().tick(state().nextPollAt - 1);

    expect(events()).toHaveLength(before);
  });

  it("does nothing when ticked while stopped", () => {
    state().stop();
    const before = events().length;
    state().tick(state().nextPollAt + 10_000);

    expect(events()).toHaveLength(before);
  });
});

describe("clock faults", () => {
  it("produces a skew close to the configured offset", () => {
    state().patchFaults({ clockOffsetMs: 45_000 });
    useBenchStore.getState().patchGateway({ scheduleMode: "interval" });
    state().start();
    runOneCycle();

    const [sample] = state().samples;
    // The wire carries whole seconds, so the sample lands within a second.
    expect(sample!.skewMs).toBeGreaterThan(44_000);
    expect(sample!.skewMs).toBeLessThanOrEqual(45_000);
  });

  it("restarts the drift origin when the rate changes", () => {
    now = T0 + 60_000;
    state().patchFaults({ clockDriftPerDayMs: 24_000 });

    expect(state().clockOrigin).toBe(now);
  });

  it("leaves the drift origin alone for an unrelated change", () => {
    state().patchFaults({ responseDelayMs: 100 });

    expect(state().clockOrigin).toBe(T0);
  });
});

describe("checksum corruption", () => {
  beforeEach(() => {
    state().patchFaults({ badChecksumProb: 1 });
    useBenchStore.getState().patchGateway({ scheduleMode: "interval", retryAttempts: 2 });
    state().start();
    runOneCycle();
  });

  it("delivers a frame that fails to decode", () => {
    const [response] = byDirection("rx");

    expect(decodeFrame(response!.bytes, "sum").ok).toBe(false);
  });

  it("raises a protocol error and retries on the same connection", () => {
    expect(byDirection("err").every((event) => event.errorCode === "invalidChecksum")).toBe(true);
    expect(state().counters.protocolErrors).toBe(3);
    expect(state().counters.retries).toBe(2);
    expect(state().counters.reconnects).toBe(0);
    expect(state().connected).toBe(true);
  });

  it("exhausts the attempts and counts one failed read", () => {
    expect(state().counters.exhaustedPolls).toBe(1);
    expect(state().counters.failedReads).toBe(1);
    expect(state().counters.successfulReads).toBe(0);
  });

  it("succeeds on a retry once the line recovers", () => {
    resetStore();
    // Each attempt draws three times: fragmentation, jitter, then corruption.
    // Holding the first three draws at 0 corrupts attempt one and nothing else.
    let draws = 0;
    jest.spyOn(Math, "random").mockImplementation(() => (draws++ < 3 ? 0 : 0.99));

    state().patchFaults({ badChecksumProb: 0.5 });
    useBenchStore.getState().patchGateway({ scheduleMode: "interval", retryAttempts: 2 });
    state().start();
    runOneCycle();

    expect(state().counters.successfulReads).toBe(1);
    expect(state().counters.retries).toBeGreaterThan(0);
    expect(state().counters.reconnects).toBe(0);
  });
});

describe("silence", () => {
  beforeEach(() => {
    state().patchFaults({ noResponse: true });
    useBenchStore.getState().patchGateway({ scheduleMode: "interval", retryAttempts: 1 });
    state().start();
    runOneCycle();
  });

  it("times out on every attempt", () => {
    expect(byDirection("err").map((event) => event.errorCode)).toEqual(["noResponse", "noResponse"]);
    expect(byDirection("rx")).toHaveLength(0);
  });

  it("counts a failed read without dropping the connection", () => {
    expect(state().counters.failedReads).toBe(1);
    expect(state().counters.reconnects).toBe(0);
    expect(state().connected).toBe(true);
  });

  it("degrades and then takes the device offline", () => {
    useBenchStore
      .getState()
      .patchGateway({ policy: { degradeAfter: 2, offlineAfter: 3, recoverAfter: 1 } });

    runOneCycle();
    expect(state().health.state).toBe("degraded");

    runOneCycle();
    expect(state().health.state).toBe("offline");
    expect(commands()).toContain("sys:device-state");
  });
});

describe("a device that drops the connection", () => {
  beforeEach(() => {
    state().patchFaults({ closeAfterRequest: true });
    useBenchStore.getState().patchGateway({ scheduleMode: "interval" });
    state().start();
    runOneCycle();
  });

  it("reports the close and reconnects", () => {
    expect(byDirection("err").map((event) => event.errorCode)).toEqual(["connectionClosed"]);
    expect(state().counters.reconnects).toBe(1);
    expect(commands()).toContain("sys:reconnect");
    expect(state().connected).toBe(true);
  });

  it("does not retry a transport failure", () => {
    expect(state().counters.retries).toBe(0);
    expect(state().counters.exhaustedPolls).toBe(0);
  });
});

describe("identity probe", () => {
  it("reads the device identity once per connection", () => {
    useBenchStore
      .getState()
      .patchGateway({ scheduleMode: "interval", intervalMs: 5000, identityProbe: true });
    state().start();
    runOneCycle();

    expect(commands().filter((entry) => entry.endsWith("read-identity"))).toEqual([
      "tx:read-identity",
      "rx:read-identity",
    ]);
    expect(state().identityRead).toEqual({
      model: "TTR20",
      serial: "SN-0000042",
      firmware: "1.2.3",
    });

    const probe = byDirection("rx").find((event) => event.command === "read-identity");
    expect(decodeFrame(probe!.bytes, "sum").frame?.data[0]).toBe(READ_IDENTITY);

    runOneCycle();
    expect(commands().filter((entry) => entry === "tx:read-identity")).toHaveLength(1);
  });

  it("is skipped when the device is silent", () => {
    state().patchFaults({ noResponse: true });
    useBenchStore.getState().patchGateway({ scheduleMode: "interval", identityProbe: true });
    state().start();
    runOneCycle();

    expect(commands()).not.toContain("tx:read-identity");
  });

  it("is skipped when disabled", () => {
    useBenchStore.getState().patchGateway({ scheduleMode: "interval", identityProbe: false });
    state().start();
    runOneCycle();

    expect(commands()).not.toContain("tx:read-identity");
  });
});

describe("settings", () => {
  it("re-plans the next poll when the schedule changes while running", () => {
    useBenchStore.getState().patchGateway({ scheduleMode: "interval", intervalMs: 5000 });
    state().start();
    runOneCycle();

    useBenchStore
      .getState()
      .patchGateway({ scheduleMode: "aligned", intervalMs: 60_000, offsetMs: 5000 });

    expect(new Date(state().nextPollAt).getUTCSeconds()).toBe(5);
  });

  it("leaves the schedule alone while stopped", () => {
    useBenchStore.getState().patchGateway({ intervalMs: 1000 });

    expect(state().nextPollAt).toBe(0);
  });

  it("stores the checksum mode and address", () => {
    state().setChecksumMode("crc16");
    state().setAdapterAddress(0x1ff);

    expect(state().checksumMode).toBe("crc16");
    expect(state().adapterAddress).toBe(0xff);
  });

  it("encodes frames in the selected checksum mode", () => {
    state().setChecksumMode("crc16");
    useBenchStore.getState().patchGateway({ scheduleMode: "interval" });
    state().start();
    runOneCycle();

    const [request] = byDirection("tx");
    expect(decodeFrame(request!.bytes, "crc16").ok).toBe(true);
    expect(decodeFrame(request!.bytes, "sum").ok).toBe(false);
  });

  it("patches the device identity", () => {
    state().setIdentity({ serial: "SN-999" });

    expect(state().identity).toEqual({ model: "TTR20", serial: "SN-999", firmware: "1.2.3" });
  });
});

describe("housekeeping", () => {
  it("clears the event log without touching the counters", () => {
    useBenchStore.getState().patchGateway({ scheduleMode: "interval" });
    state().start();
    runOneCycle();

    state().clearEvents();

    expect(events()).toEqual([]);
    expect(state().counters.successfulReads).toBe(1);
  });

  it("resets everything the run produced", () => {
    useBenchStore.getState().patchGateway({ scheduleMode: "interval" });
    state().start();
    runOneCycle();

    state().reset();

    expect(state().running).toBe(false);
    expect(state().events).toEqual([]);
    expect(state().samples).toEqual([]);
    expect(state().counters.successfulReads).toBe(0);
    expect(state().health).toEqual(EMPTY_HEALTH);
  });

  it("bounds the event log", () => {
    useBenchStore.getState().patchGateway({ scheduleMode: "interval", intervalMs: 1000 });
    state().start();
    for (let index = 0; index < 300; index += 1) runOneCycle();

    expect(events().length).toBeLessThanOrEqual(400);
    expect(state().samples.length).toBeLessThanOrEqual(120);
  });
});

describe("selectors", () => {
  it("counts the active faults", () => {
    expect(selectActiveFaultCount(state())).toBe(0);

    state().patchFaults({
      responseDelayMs: 100,
      badChecksumProb: 0.5,
      noResponse: true,
      clockOffsetMs: -1000,
    });

    expect(selectActiveFaultCount(state())).toBe(4);
  });

  it("returns the last frame as hex", () => {
    expect(selectLastFrameHex(state())).toBe("");

    useBenchStore.getState().patchGateway({ scheduleMode: "interval" });
    state().start();
    runOneCycle();

    const last = [...events()].reverse().find((event) => event.bytes.length > 0)!;
    expect(selectLastFrameHex(state())).toBe(formatHex(last.bytes));
  });

  it("reports availability from the outcome window", () => {
    expect(selectAvailability(state())).toBe(0);
  });
});

describe("a device slower than the request timeout", () => {
  it("times out rather than accepting a late answer", () => {
    state().patchFaults({ responseDelayMs: 5000 });
    useBenchStore
      .getState()
      .patchGateway({ scheduleMode: "interval", requestTimeoutMs: 500, retryAttempts: 1 });
    state().start();
    runOneCycle();

    expect(byDirection("err").map((event) => event.errorCode)).toEqual(["timeout", "timeout"]);
    expect(byDirection("rx")).toHaveLength(0);
    expect(state().counters.failedReads).toBe(1);
    expect(state().counters.reconnects).toBe(0);
  });

  it("accepts the answer once the timeout is raised above it", () => {
    state().patchFaults({ responseDelayMs: 400 });
    useBenchStore
      .getState()
      .patchGateway({ scheduleMode: "interval", requestTimeoutMs: 1500, retryAttempts: 0 });
    state().start();
    runOneCycle();

    expect(state().counters.successfulReads).toBe(1);
    expect(byDirection("rx")[0]?.latencyMs).toBeGreaterThanOrEqual(400);
  });
});

describe("fragmentation", () => {
  it("adds the fragment delay to the round trip", () => {
    state().patchFaults({ fragmentProb: 1, fragmentDelayMs: 300 });
    useBenchStore.getState().patchGateway({ scheduleMode: "interval", requestTimeoutMs: 1500 });
    state().start();
    runOneCycle();

    expect(byDirection("rx")[0]?.latencyMs).toBeGreaterThanOrEqual(300);
    expect(state().counters.successfulReads).toBe(1);
  });
});
