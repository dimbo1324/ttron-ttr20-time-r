import { buildClockReport, recordOutcome, sampleSkew, EMPTY_HEALTH } from "@/lib/bench/domain";
import { useBenchStore } from "@/stores/bench-store";
import { resetBenchStore } from "@/test/utils";

import {
  benchCounters,
  benchFaults,
  benchHealth,
  benchLimits,
  benchSchedule,
  benchTelemetry,
  benchTicks,
} from "./from-bench";
import { activeFaultCount } from "./types";

/**
 * The bench, in the console's vocabulary.
 *
 * Mostly a projection, so the tests concentrate on the two places it does real
 * work — folding the outcome window into an availability figure and latency
 * percentiles, which the live source reports directly — and on the ticks,
 * where the bench knows something the live source cannot.
 */

const state = () => useBenchStore.getState();

const policy = { degradeAfter: 3, offlineAfter: 10, recoverAfter: 2 };

beforeEach(() => {
  resetBenchStore();
});

describe("health", () => {
  it("reports an untouched bench as having measured nothing", () => {
    const health = benchHealth(state());

    expect(health.state).toBe("unknown");
    expect(health.window).toEqual([]);
    expect(health.windowSamples).toBe(0);
    expect(health.availability).toBe(0);
  });

  it("folds the outcome window into an availability figure", () => {
    let snapshot = EMPTY_HEALTH;
    snapshot = recordOutcome(snapshot, true, policy, 10);
    snapshot = recordOutcome(snapshot, true, policy, 30);
    snapshot = recordOutcome(snapshot, false, policy);
    snapshot = recordOutcome(snapshot, true, policy, 20);
    resetBenchStore({ health: snapshot });

    const health = benchHealth(state());

    expect(health.window).toEqual([true, true, false, true]);
    expect(health.availability).toBe(0.75);
    expect(health.windowSamples).toBe(4);
  });

  it("computes the percentiles the live source reports directly", () => {
    let snapshot = EMPTY_HEALTH;
    for (const latency of [10, 20, 30, 40, 100]) {
      snapshot = recordOutcome(snapshot, true, policy, latency);
    }
    resetBenchStore({ health: snapshot });

    const health = benchHealth(state());

    // Folded here rather than in the dashboard, which is what used to make
    // the dashboard bench-only.
    expect(health.latencyP50Ms).toBe(30);
    expect(health.latencyP95Ms).toBe(100);
  });

  it("carries the consecutive counts the hysteresis rule is written against", () => {
    let snapshot = recordOutcome(EMPTY_HEALTH, true, policy, 5);
    snapshot = recordOutcome(snapshot, false, policy);
    snapshot = recordOutcome(snapshot, false, policy);
    resetBenchStore({ health: snapshot });

    const health = benchHealth(state());

    expect(health.consecutiveFailures).toBe(2);
    expect(health.consecutiveSuccesses).toBe(0);
  });
});

describe("counters and settings", () => {
  it("maps the six counters the console shows", () => {
    resetBenchStore({
      counters: {
        successfulReads: 5,
        failedReads: 4,
        reconnects: 3,
        retries: 2,
        protocolErrors: 1,
        exhaustedPolls: 9,
        connections: 6,
      },
    });

    expect(benchCounters(state())).toEqual({
      successfulReads: 5,
      failedReads: 4,
      retries: 2,
      protocolErrors: 1,
      reconnects: 3,
      connections: 6,
    });
  });

  it("reads the schedule and the next poll instant", () => {
    resetBenchStore({ nextPollAt: 1_700_000_000_000 });

    expect(benchSchedule(state())).toEqual({
      mode: "aligned",
      intervalMs: 5000,
      offsetMs: 0,
      nextPollAt: 1_700_000_000_000,
    });
  });

  it("reports no connect timeout, because the bench dials nothing", () => {
    const limits = benchLimits(state());

    expect(limits.requestTimeoutMs).toBe(1500);
    expect(limits.retryAttempts).toBe(2);
    expect(limits.connectTimeoutMs).toBe(0);
  });

  it("reports the clock faults as numbers, not as absent", () => {
    resetBenchStore({
      faults: { ...state().faults, clockOffsetMs: 0, clockDriftPerDayMs: -4000 },
    });

    const faults = benchFaults(state());

    // Zero here means "the clock is correct", which is a real reading. Null
    // is reserved for a source that cannot move the clock at all.
    expect(faults.clockOffsetMs).toBe(0);
    expect(faults.clockDriftPerDayMs).toBe(-4000);
  });
});

describe("schedule ticks", () => {
  const tx = (at: number, attempt: number, command = "read-time") => ({
    id: at,
    at,
    direction: "tx" as const,
    source: "gateway" as const,
    command,
    bytes: [],
    cycle: 1,
    attempt,
  });

  it("plots only the request that opened each cycle", () => {
    resetBenchStore({
      events: [tx(1000, 0), tx(1200, 1), tx(1400, 2), tx(6000, 0)],
    });

    // Retries are requests, but they are not schedule events, and including
    // them would make a perfectly regular schedule look ragged.
    expect(benchTicks(state())).toEqual([1000, 6000]);
  });

  it("ignores the identity probe", () => {
    resetBenchStore({ events: [tx(1000, 0, "read-identity"), tx(1010, 0)] });

    expect(benchTicks(state())).toEqual([1010]);
  });
});

describe("the whole projection", () => {
  it("describes an idle bench", () => {
    const clock = buildClockReport([], { warnMs: 2000, criticalMs: 30_000 });
    const telemetry = benchTelemetry(state(), clock);

    expect(telemetry.source).toBe("bench");
    // The bench runs in this tab; there is nothing it could fail to reach.
    expect(telemetry.link).toBe("ready");
    expect(telemetry.error).toBeNull();
    expect(telemetry.editable).toBe(true);
    expect(telemetry.target).toBeNull();
    // A simulated device is a device, not a fleet.
    expect(telemetry.fleet).toBeNull();
  });

  it("passes the clock report through rather than recomputing it", () => {
    const samples = [sampleSkew(0, 10, 4000), sampleSkew(1000, 1010, 5100)];
    resetBenchStore({ samples });
    const clock = buildClockReport(samples, { warnMs: 2000, criticalMs: 30_000 });

    const telemetry = benchTelemetry(state(), clock);

    // The hook that owns this projection memoises the report; recomputing it
    // here would hand React a new object on every render.
    expect(telemetry.clock).toBe(clock);
    expect(telemetry.samples).toBe(samples);
  });

  it("carries the faults and their count", () => {
    resetBenchStore({
      faults: { ...state().faults, noResponse: true, responseDelayMs: 500 },
    });

    const telemetry = benchTelemetry(state(), buildClockReport([], state().gateway.thresholds));

    expect(telemetry.faults?.noResponse).toBe(true);
    expect(activeFaultCount(telemetry.faults)).toBe(2);
  });

  it("counts no faults at all when there is no device to fault", () => {
    expect(activeFaultCount(null)).toBe(0);
  });

  it("does not count a clock the source cannot move", () => {
    expect(
      activeFaultCount({
        responseDelayMs: 0,
        badChecksumProb: 0,
        fragmentProb: 0,
        fragmentDelayMs: 40,
        noResponse: false,
        closeAfterRequest: false,
        clockOffsetMs: null,
        clockDriftPerDayMs: null,
      }),
    ).toBe(0);
  });
});
