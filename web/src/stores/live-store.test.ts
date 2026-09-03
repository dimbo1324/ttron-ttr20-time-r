import { api, ApiError } from "@/lib/api/client";
import {
  eventFixture,
  faultModeFixture,
  fleetFixture,
  gatewayStatusFixture,
  historyFixture,
  resetLiveStore,
} from "@/test/utils";

import { liveTelemetry, useLiveStore } from "./live-store";

/**
 * The live source, with the API client stubbed.
 *
 * Stubbing the client rather than `fetch` keeps these tests about the store's
 * own decisions — what a partial failure does to the view, what an optimistic
 * write rolls back to — while the client's own tests cover the wire.
 */
jest.mock("@/lib/api/client", () => {
  const actual = jest.requireActual<typeof import("@/lib/api/client")>("@/lib/api/client");
  return {
    ...actual,
    api: {
      gatewayStatus: jest.fn(),
      gatewayEvents: jest.fn(),
      gatewayHistory: jest.fn(),
      gatewayFleet: jest.fn(),
      startPolling: jest.fn(),
      stopPolling: jest.fn(),
      faultMode: jest.fn(),
      setFaultMode: jest.fn(),
      emulatorStatus: jest.fn(),
    },
  };
});

const mocked = api as jest.Mocked<typeof api>;
const parsedStatus = () => gatewayStatusFixture() as never;

function allAnswer() {
  mocked.gatewayStatus.mockResolvedValue(parsedStatus());
  mocked.gatewayEvents.mockResolvedValue([eventFixture()] as never);
  mocked.gatewayHistory.mockResolvedValue(historyFixture() as never);
  mocked.gatewayFleet.mockResolvedValue(fleetFixture() as never);
}

beforeEach(() => {
  resetLiveStore();
});

describe("refresh", () => {
  it("commits everything from one instant together", async () => {
    allAnswer();

    await useLiveStore.getState().refresh();
    const state = useLiveStore.getState();

    expect(state.link).toBe("ready");
    expect(state.status).not.toBeNull();
    expect(state.history).not.toBeNull();
    expect(state.fleet).not.toBeNull();
    expect(state.events).toHaveLength(1);
  });

  it("goes unreachable when the status call finds nothing running", async () => {
    allAnswer();
    mocked.gatewayStatus.mockRejectedValue(new ApiError("Failed to fetch", 0, "NO_ANSWER"));

    await useLiveStore.getState().refresh();

    expect(useLiveStore.getState().link).toBe("unreachable");
    expect(useLiveStore.getState().status).toBeNull();
  });

  it("stays up when the API answers with an error of its own", async () => {
    allAnswer();
    mocked.gatewayStatus.mockRejectedValue(
      new ApiError("gateway gRPC service is unavailable", 503, "GATEWAY_UNAVAILABLE"),
    );

    await useLiveStore.getState().refresh();

    // The API is reachable and reporting on the gateway behind it. Calling
    // that "unreachable" would send the reader after the wrong process.
    expect(useLiveStore.getState().link).toBe("ready");
    expect(useLiveStore.getState().error).toBe("gateway gRPC service is unavailable");
  });

  it("keeps the last history when a gateway too old to answer 404s", async () => {
    allAnswer();
    await useLiveStore.getState().refresh();

    mocked.gatewayHistory.mockRejectedValue(new ApiError("HTTP 404", 404, "NO_ANSWER"));
    mocked.gatewayFleet.mockRejectedValue(new ApiError("HTTP 404", 404, "NO_ANSWER"));
    await useLiveStore.getState().refresh();

    const state = useLiveStore.getState();
    // Only the status decides whether the link is up; the newer endpoints
    // degrade to what they last held rather than blanking the page.
    expect(state.link).toBe("ready");
    expect(state.history).not.toBeNull();
    expect(state.fleet).not.toBeNull();
  });

  it("keeps the last log when the events call fails on its own", async () => {
    allAnswer();
    await useLiveStore.getState().refresh();

    mocked.gatewayEvents.mockRejectedValue(new ApiError("HTTP 500", 500, "NO_ANSWER"));
    await useLiveStore.getState().refresh();

    expect(useLiveStore.getState().events).toHaveLength(1);
  });

  it("reports a gateway error the status itself carries", async () => {
    allAnswer();
    mocked.gatewayStatus.mockResolvedValue(
      gatewayStatusFixture({ lastError: "dial tcp: connection refused" }) as never,
    );

    await useLiveStore.getState().refresh();

    expect(useLiveStore.getState().error).toBe("dial tcp: connection refused");
  });

  it("caps the log the same way the bench engine does", async () => {
    allAnswer();
    mocked.gatewayEvents.mockResolvedValue(
      Array.from({ length: 500 }, (_, index) =>
        eventFixture({ id: index + 1, timestamp: new Date(index * 1000).toISOString() }),
      ) as never,
    );

    await useLiveStore.getState().refresh();

    // The newest are kept: an unbounded log is a memory leak on a bench left
    // running overnight.
    const events = useLiveStore.getState().events;
    expect(events).toHaveLength(400);
    expect(events[events.length - 1]!.id).toBe(500);
  });
});

describe("start and stop", () => {
  it("adopts the status the command answered with", async () => {
    mocked.startPolling.mockResolvedValue(parsedStatus());

    await useLiveStore.getState().start();

    expect(useLiveStore.getState().status).not.toBeNull();
    expect(useLiveStore.getState().link).toBe("ready");
    expect(useLiveStore.getState().busy).toBe(false);
  });

  it("clears a stale error on a successful command", async () => {
    useLiveStore.setState({ error: "something earlier" });
    mocked.stopPolling.mockResolvedValue(parsedStatus());

    await useLiveStore.getState().stop();

    expect(useLiveStore.getState().error).toBeNull();
  });

  it("reports a command that could not be delivered", async () => {
    mocked.startPolling.mockRejectedValue(new ApiError("Failed to fetch", 0, "NO_ANSWER"));

    await useLiveStore.getState().start();

    expect(useLiveStore.getState().link).toBe("unreachable");
    expect(useLiveStore.getState().busy).toBe(false);
  });

  it("clears busy even when the command throws something unexpected", async () => {
    mocked.stopPolling.mockRejectedValue("not an Error at all");

    await useLiveStore.getState().stop();

    expect(useLiveStore.getState().busy).toBe(false);
    expect(useLiveStore.getState().error).toBe("unknown error");
  });

  it("keeps the message of a plain Error", async () => {
    // Not every failure comes from the client; a bug in this store would
    // arrive as an ordinary Error, and its message is still the best thing
    // to put on screen.
    mocked.startPolling.mockRejectedValue(new Error("Cannot read properties of undefined"));

    await useLiveStore.getState().start();

    expect(useLiveStore.getState().link).toBe("unreachable");
    expect(useLiveStore.getState().error).toBe("Cannot read properties of undefined");
  });
});

describe("fault mode", () => {
  it("reads the emulator's current faults", async () => {
    mocked.faultMode.mockResolvedValue(faultModeFixture({ responseDelayMs: 120 }) as never);

    await useLiveStore.getState().loadFaults();

    expect(useLiveStore.getState().faults?.responseDelayMs).toBe(120);
  });

  it("stays quiet when the emulator is not running", async () => {
    mocked.faultMode.mockRejectedValue(new ApiError("Failed to fetch", 0, "NO_ANSWER"));

    await useLiveStore.getState().loadFaults();

    // The emulator is a separate process from the gateway; its absence must
    // not be reported as the gateway being down.
    expect(useLiveStore.getState().faults).toBeNull();
    expect(useLiveStore.getState().link).toBe("connecting");
  });

  it("shows the change immediately, then settles on what was accepted", async () => {
    mocked.faultMode.mockResolvedValue(faultModeFixture() as never);
    await useLiveStore.getState().loadFaults();

    // The emulator raises its own flag when a probability is set; the store
    // takes whatever comes back rather than what it sent.
    mocked.setFaultMode.mockResolvedValue(
      faultModeFixture({ corruptChecksum: true, corruptChecksumProbability: 0.35 }) as never,
    );
    await useLiveStore.getState().setFaults({ corruptChecksumProbability: 0.35 });

    expect(mocked.setFaultMode).toHaveBeenCalledWith(
      expect.objectContaining({ corruptChecksumProbability: 0.35 }),
    );
    expect(useLiveStore.getState().faults?.corruptChecksum).toBe(true);
  });

  it("puts the slider back when the emulator refuses", async () => {
    mocked.faultMode.mockResolvedValue(faultModeFixture({ responseDelayMs: 50 }) as never);
    await useLiveStore.getState().loadFaults();

    mocked.setFaultMode.mockRejectedValue(new ApiError("Failed to fetch", 0, "NO_ANSWER"));
    await useLiveStore.getState().setFaults({ responseDelayMs: 900 });

    // Leaving 900 on screen would tell the operator a fault is active on a
    // device that never received it.
    expect(useLiveStore.getState().faults?.responseDelayMs).toBe(50);
    expect(useLiveStore.getState().link).toBe("unreachable");
  });

  it("does nothing before the emulator has been read", async () => {
    await useLiveStore.getState().setFaults({ noResponse: true });

    expect(mocked.setFaultMode).not.toHaveBeenCalled();
  });
});

describe("telemetry projection", () => {
  it("renders a full, empty reading before the first refresh", () => {
    const telemetry = liveTelemetry(useLiveStore.getState());

    // A nullable source would put a "is it loaded yet" branch into every
    // panel; an empty reading with a link state says the same thing once.
    expect(telemetry.source).toBe("live");
    expect(telemetry.link).toBe("connecting");
    expect(telemetry.clock.state).toBe("unknown");
    expect(telemetry.health.state).toBe("unknown");
    expect(telemetry.counters.successfulReads).toBe(0);
    expect(telemetry.running).toBe(false);
    expect(telemetry.fleet).toBeNull();
  });

  it("drops every current reading when the link goes down, and keeps the log", async () => {
    allAnswer();
    await useLiveStore.getState().refresh();

    mocked.gatewayStatus.mockRejectedValue(new ApiError("Failed to fetch", 0, "NO_ANSWER"));
    await useLiveStore.getState().refresh();

    const telemetry = liveTelemetry(useLiveStore.getState());
    // A stale fleet table beside an "unreachable" notice reads as current.
    expect(telemetry.fleet).toBeNull();
    expect(telemetry.clock.state).toBe("unknown");
    // The log rows carry their own timestamps and cannot be mistaken for now.
    expect(telemetry.events).toHaveLength(1);
  });

  it("projects a refreshed store into readable telemetry", async () => {
    allAnswer();
    mocked.faultMode.mockResolvedValue(faultModeFixture() as never);
    await useLiveStore.getState().refresh();
    await useLiveStore.getState().loadFaults();

    const telemetry = liveTelemetry(useLiveStore.getState());

    expect(telemetry.running).toBe(true);
    expect(telemetry.connected).toBe(true);
    expect(telemetry.target).toBe("127.0.0.1:9000");
    expect(telemetry.clock.medianMs).toBe(-2400);
    expect(telemetry.thresholds).toEqual({ warnMs: 2000, criticalMs: 30_000 });
    expect(telemetry.policy).toEqual({ degradeAfter: 3, offlineAfter: 10, recoverAfter: 2 });
    expect(telemetry.limits).toEqual({
      requestTimeoutMs: 1500,
      connectTimeoutMs: 2000,
      retryAttempts: 2,
    });
    expect(telemetry.identity).toEqual({ model: "TTR20", serial: "SN-42", firmware: "1.2.3" });
    expect(telemetry.lastDeviceTime).toBe(Date.UTC(2026, 8, 2, 11, 59, 58));
  });

  it("never lets the live source be edited", async () => {
    allAnswer();
    await useLiveStore.getState().refresh();

    // Interval, thresholds and the health policy are the gateway process's
    // configuration; the control plane has no setter for any of them.
    expect(liveTelemetry(useLiveStore.getState()).editable).toBe(false);
  });

  it("reports a stopped gateway as not running", async () => {
    allAnswer();
    mocked.gatewayStatus.mockResolvedValue(
      gatewayStatusFixture({ state: "stopped", connected: false }) as never,
    );
    await useLiveStore.getState().refresh();

    const telemetry = liveTelemetry(useLiveStore.getState());
    expect(telemetry.running).toBe(false);
    expect(telemetry.connected).toBe(false);
  });

  it("draws nothing rather than guessing when the history is missing", async () => {
    allAnswer();
    mocked.gatewayHistory.mockRejectedValue(new ApiError("HTTP 404", 404, "NO_ANSWER"));
    await useLiveStore.getState().refresh();

    const telemetry = liveTelemetry(useLiveStore.getState());
    expect(telemetry.samples).toEqual([]);
    expect(telemetry.health.window).toEqual([]);
    // The aggregates still come from the status, which every gateway sends.
    expect(telemetry.health.availability).toBe(0.875);
  });

  it("reads the clock flags the way the emulator means them", async () => {
    // The emulator raises the flag whenever the probability is above zero, so
    // the flag alone says nothing; only a flag with no probability is
    // "every frame". Reading it otherwise snapped a 35% slider to 100%.
    mocked.faultMode.mockResolvedValue(
      faultModeFixture({ corruptChecksum: true, corruptChecksumProbability: 0.35 }) as never,
    );
    await useLiveStore.getState().loadFaults();

    expect(liveTelemetry(useLiveStore.getState()).faults?.badChecksumProb).toBe(0.35);
  });

  it("reads a flag with no probability as every frame", async () => {
    mocked.faultMode.mockResolvedValue(
      faultModeFixture({
        corruptChecksum: true,
        corruptChecksumProbability: 0,
        fragmentResponse: true,
        fragmentProbability: 0,
      }) as never,
    );
    await useLiveStore.getState().loadFaults();

    const faults = liveTelemetry(useLiveStore.getState()).faults;
    expect(faults?.badChecksumProb).toBe(1);
    expect(faults?.fragmentProb).toBe(1);
  });

  it("has no fleet before one has been fetched", async () => {
    allAnswer();
    mocked.gatewayFleet.mockRejectedValue(new ApiError("HTTP 404", 404, "NO_ANSWER"));
    await useLiveStore.getState().refresh();

    // A gateway too old to answer /fleet still reports its own status; the
    // table is simply absent rather than showing one invented row.
    expect(liveTelemetry(useLiveStore.getState()).fleet).toBeNull();
  });

  it("says never rather than the epoch for a gateway that has not read yet", async () => {
    allAnswer();
    mocked.gatewayStatus.mockResolvedValue(
      gatewayStatusFixture({
        targetAddr: "",
        lastDeviceTime: undefined,
        lastSuccessfulReadTime: undefined,
      }) as never,
    );
    await useLiveStore.getState().refresh();

    const telemetry = liveTelemetry(useLiveStore.getState());
    // Zero and null read as "—" at the call site; 1970 would read as a
    // device that answered fifty years ago.
    expect(telemetry.lastDeviceTime).toBeNull();
    expect(telemetry.lastCycleAt).toBe(0);
    expect(telemetry.target).toBeNull();
  });

  it("has no clock fault to report, because the Go emulator cannot inject one", async () => {
    mocked.faultMode.mockResolvedValue(faultModeFixture() as never);
    await useLiveStore.getState().loadFaults();

    const faults = liveTelemetry(useLiveStore.getState()).faults;
    // Null rather than zero: zero would claim the clock had been checked and
    // found correct.
    expect(faults?.clockOffsetMs).toBeNull();
    expect(faults?.clockDriftPerDayMs).toBeNull();
  });
});
