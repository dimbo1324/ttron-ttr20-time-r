import {
  availability,
  buildClockReport,
  classifySkew,
  computeDrift,
  EMPTY_HEALTH,
  MIN_DRIFT_SAMPLES,
  nextPollAt,
  percentile,
  recordOutcome,
  sampleSkew,
  type ClockThresholds,
  type HealthPolicy,
  type SkewSample,
} from "./domain";

const THRESHOLDS: ClockThresholds = { warnMs: 2000, criticalMs: 30_000 };
const POLICY: HealthPolicy = { degradeAfter: 3, offlineAfter: 5, recoverAfter: 2 };

/** 2026-09-02T12:00:00Z, so aligned maths is checked against a real instant. */
const ORIGIN = Date.UTC(2026, 8, 2, 12, 0, 0);

describe("nextPollAt in interval mode", () => {
  it("counts from now", () => {
    expect(nextPollAt(ORIGIN, "interval", 5000, 0)).toBe(ORIGIN + 5000);
  });

  it("ignores the offset", () => {
    expect(nextPollAt(ORIGIN, "interval", 5000, 3000)).toBe(ORIGIN + 5000);
  });

  it("is always in the future", () => {
    let now = ORIGIN;
    for (let index = 0; index < 5; index += 1) {
      const next = nextPollAt(now, "interval", 1000, 0);
      expect(next).toBeGreaterThan(now);
      now = next;
    }
  });
});

describe("nextPollAt in aligned mode", () => {
  it("lands on the fifth second of every minute", () => {
    const base = Date.UTC(2026, 8, 2, 12, 3, 0);
    const cases = [
      { now: base + 4000, want: base + 5000 },
      { now: base + 5000, want: base + 65_000 },
      { now: base + 7250, want: base + 65_000 },
      { now: base + 59_500, want: base + 65_000 },
    ];

    for (const { now, want } of cases) {
      expect(nextPollAt(now, "aligned", 60_000, 5000)).toBe(want);
      expect(new Date(nextPollAt(now, "aligned", 60_000, 5000)).getUTCSeconds()).toBe(5);
    }
  });

  it("snaps to five-second boundaries with no offset", () => {
    const base = Date.UTC(2026, 8, 2, 12, 3, 0);

    expect(nextPollAt(base + 7250, "aligned", 5000, 0)).toBe(base + 10_000);
    expect(nextPollAt(base + 10_000, "aligned", 5000, 0)).toBe(base + 15_000);
  });

  it("keeps its phase across repeated calls", () => {
    let now = ORIGIN + 137;
    for (let index = 0; index < 12; index += 1) {
      const next = nextPollAt(now, "aligned", 60_000, 5000);
      expect(new Date(next).getUTCSeconds()).toBe(5);
      expect(next).toBeGreaterThan(now);
      // Resume from just after the tick, the way a real cycle would.
      now = next + 17;
    }
  });

  it("handles instants before the Unix epoch", () => {
    const before = Date.UTC(1969, 11, 31, 23, 58, 30);
    const next = nextPollAt(before, "aligned", 60_000, 5000);

    expect(next).toBeGreaterThan(before);
    expect(new Date(next).getUTCSeconds()).toBe(5);
  });
});

describe("classifySkew", () => {
  it.each([
    { skew: 0, want: "ok" },
    { skew: 1999, want: "ok" },
    { skew: 2000, want: "warn" },
    { skew: -2000, want: "warn" },
    { skew: 10_000, want: "warn" },
    { skew: 30_000, want: "critical" },
    { skew: -45_000, want: "critical" },
  ])("classifies $skew as $want", ({ skew, want }) => {
    expect(classifySkew(skew, THRESHOLDS)).toBe(want);
  });
});

describe("sampleSkew", () => {
  it("compensates half the round trip", () => {
    const sample = sampleSkew(1000, 1200, 1100);

    expect(sample.roundTripMs).toBe(200);
    expect(sample.at).toBe(1100);
    expect(sample.skewMs).toBe(0);
  });

  it("keeps the sign of a device running fast", () => {
    expect(sampleSkew(1000, 1040, 6020).skewMs).toBe(5000);
  });

  it("keeps the sign of a device running slow", () => {
    expect(sampleSkew(1000, 1040, -3980).skewMs).toBe(-5000);
  });

  it("clamps a negative round trip to zero", () => {
    const sample = sampleSkew(2000, 1000, 2000);

    expect(sample.roundTripMs).toBe(0);
    expect(sample.at).toBe(2000);
  });
});

function series(count: number, step: number, skewStep: number): SkewSample[] {
  return Array.from({ length: count }, (_, index) => ({
    at: ORIGIN + index * step,
    skewMs: index * skewStep,
    roundTripMs: 10,
  }));
}

describe("computeDrift", () => {
  it("needs a minimum number of samples", () => {
    for (let count = 0; count < MIN_DRIFT_SAMPLES; count += 1) {
      expect(computeDrift(series(count, 3_600_000, 1000)).determined).toBe(false);
    }
  });

  it("measures a linear gain in ms per day", () => {
    // One second gained per hour is 24 seconds per day.
    const drift = computeDrift(series(4, 3_600_000, 1000));

    expect(drift.determined).toBe(true);
    expect(drift.perDayMs).toBeCloseTo(24_000, 3);
    expect(drift.fit).toBeCloseTo(1, 6);
  });

  it("measures a linear loss", () => {
    const drift = computeDrift(series(4, 1_800_000, -1000));

    expect(drift.perDayMs).toBeCloseTo(-48_000, 3);
  });

  it("reports a flat series as zero drift", () => {
    const samples = series(5, 3_600_000, 0);
    const drift = computeDrift(samples);

    expect(drift.determined).toBe(true);
    expect(drift.perDayMs).toBe(0);
    expect(drift.fit).toBe(0);
  });

  it("cannot determine drift without time spread", () => {
    const samples: SkewSample[] = [
      { at: ORIGIN, skewMs: 1000, roundTripMs: 10 },
      { at: ORIGIN, skewMs: 2000, roundTripMs: 10 },
      { at: ORIGIN, skewMs: 3000, roundTripMs: 10 },
    ];

    expect(computeDrift(samples).determined).toBe(false);
  });

  it("reports a poor fit for a noisy series", () => {
    const samples: SkewSample[] = [0, 5000, -5000, 10_000, -10_000].map((skewMs, index) => ({
      at: ORIGIN + index * 3_600_000,
      skewMs,
      roundTripMs: 10,
    }));

    expect(computeDrift(samples).fit).toBeLessThan(0.5);
  });
});

describe("buildClockReport", () => {
  it("is unknown with no samples", () => {
    const report = buildClockReport([], THRESHOLDS);

    expect(report.state).toBe("unknown");
    expect(report.samples).toBe(0);
    expect(report.driftDetermined).toBe(false);
  });

  it("classifies on the latest sample before the window fills", () => {
    const report = buildClockReport(
      [{ at: ORIGIN, skewMs: 5000, roundTripMs: 12 }],
      THRESHOLDS,
    );

    expect(report.state).toBe("warn");
    expect(report.skewMs).toBe(5000);
    expect(report.roundTripMs).toBe(12);
  });

  it("classifies on the median once there are enough samples", () => {
    const samples: SkewSample[] = [0, 0, 0, 45_000].map((skewMs, index) => ({
      at: ORIGIN + index * 60_000,
      skewMs,
      roundTripMs: 10,
    }));
    const report = buildClockReport(samples, THRESHOLDS);

    expect(report.state).toBe("ok");
    expect(report.skewMs).toBe(45_000);
    expect(report.maxMs).toBe(45_000);
    expect(report.minMs).toBe(0);
  });

  it("escalates on a sustained skew", () => {
    const samples: SkewSample[] = Array.from({ length: 6 }, (_, index) => ({
      at: ORIGIN + index * 60_000,
      skewMs: 40_000,
      roundTripMs: 10,
    }));
    const report = buildClockReport(samples, THRESHOLDS);

    expect(report.state).toBe("critical");
    expect(report.medianMs).toBe(40_000);
  });

  it("averages the middle pair for an even sample count", () => {
    const samples: SkewSample[] = [2000, 4000, 6000, 8000].map((skewMs, index) => ({
      at: ORIGIN + index * 60_000,
      skewMs,
      roundTripMs: 10,
    }));

    expect(buildClockReport(samples, THRESHOLDS).medianMs).toBe(5000);
  });
});

describe("recordOutcome", () => {
  const fail = (snapshot = EMPTY_HEALTH, times = 1) => {
    let current = snapshot;
    for (let index = 0; index < times; index += 1) current = recordOutcome(current, false, POLICY);
    return current;
  };

  it("starts unknown", () => {
    expect(EMPTY_HEALTH.state).toBe("unknown");
  });

  it("goes online on the first success", () => {
    const snapshot = recordOutcome(EMPTY_HEALTH, true, POLICY, 12);

    expect(snapshot.state).toBe("online");
    expect(snapshot.successes).toBe(1);
    expect(snapshot.latencies).toEqual([12]);
  });

  it("stays unknown while it has never been reached", () => {
    expect(fail(EMPTY_HEALTH, 2).state).toBe("unknown");
    expect(fail(EMPTY_HEALTH, 3).state).toBe("degraded");
  });

  it("degrades and then goes offline with hysteresis", () => {
    let snapshot = recordOutcome(EMPTY_HEALTH, true, POLICY, 10);

    snapshot = fail(snapshot, 2);
    expect(snapshot.state).toBe("online");

    snapshot = fail(snapshot, 1);
    expect(snapshot.state).toBe("degraded");

    snapshot = fail(snapshot, 2);
    expect(snapshot.state).toBe("offline");
    expect(snapshot.consecutiveFailures).toBe(5);
  });

  it("needs consecutive successes to recover", () => {
    let snapshot = fail(EMPTY_HEALTH, 5);
    expect(snapshot.state).toBe("offline");

    snapshot = recordOutcome(snapshot, true, POLICY, 10);
    expect(snapshot.state).toBe("offline");

    snapshot = recordOutcome(snapshot, true, POLICY, 10);
    expect(snapshot.state).toBe("online");
  });

  it("resets the failure streak on a success", () => {
    let snapshot = recordOutcome(EMPTY_HEALTH, true, POLICY, 10);
    snapshot = fail(snapshot, 2);
    snapshot = recordOutcome(snapshot, true, POLICY, 10);

    expect(snapshot.consecutiveFailures).toBe(0);

    snapshot = fail(snapshot, 2);
    expect(snapshot.state).toBe("online");
  });

  it("does not record a latency for a failure", () => {
    const snapshot = recordOutcome(EMPTY_HEALTH, false, POLICY);

    expect(snapshot.latencies).toEqual([]);
    expect(snapshot.failures).toBe(1);
  });

  it("bounds the outcome window", () => {
    let snapshot = EMPTY_HEALTH;
    for (let index = 0; index < 200; index += 1) {
      snapshot = recordOutcome(snapshot, true, POLICY, index);
    }

    expect(snapshot.window.length).toBeLessThanOrEqual(60);
    expect(snapshot.latencies.length).toBeLessThanOrEqual(60);
    expect(snapshot.successes).toBe(200);
  });
});

describe("availability", () => {
  it("is zero for an empty window", () => {
    expect(availability(EMPTY_HEALTH)).toBe(0);
  });

  it.each([
    { window: [true, true, true], want: 1 },
    { window: [false, false], want: 0 },
    { window: [true, false, true, false], want: 0.5 },
  ])("is $want for $window", ({ window, want }) => {
    expect(availability({ ...EMPTY_HEALTH, window })).toBe(want);
  });
});

describe("percentile", () => {
  const values = [10, 20, 30, 40, 50];

  it.each([
    { quantile: 0, want: 10 },
    { quantile: 0.5, want: 30 },
    { quantile: 0.95, want: 50 },
    { quantile: 1, want: 50 },
  ])("p$quantile is $want", ({ quantile, want }) => {
    expect(percentile(values, quantile)).toBe(want);
  });

  it("is zero for no values", () => {
    expect(percentile([], 0.5)).toBe(0);
  });

  it("clamps out-of-range quantiles", () => {
    expect(percentile(values, -1)).toBe(10);
    expect(percentile(values, 5)).toBe(50);
  });

  it("does not mutate its input", () => {
    const unsorted = [30, 10, 20];
    percentile(unsorted, 0.5);

    expect(unsorted).toEqual([30, 10, 20]);
  });
});
