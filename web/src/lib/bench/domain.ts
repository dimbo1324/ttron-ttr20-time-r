/**
 * The domain model the bench simulates, ported from the Go services.
 *
 * These are the three things the gateway actually reasons about — when to
 * poll, whether the device clock is drifting, and whether the device is still
 * there — and the console has to show all three whether or not a backend is
 * running. Keeping them as small pure functions (rather than folding them into
 * the store) is what lets the reference pages explain them: the same
 * `classifySkew` that drives the dashboard badge is the one the teaching page
 * walks through.
 */

/* ------------------------------------------------------------------ schedule */

export type ScheduleMode = "interval" | "aligned";

/**
 * Next poll instant.
 *
 * `interval` counts from now, so its phase is whatever the connection happened
 * to be. `aligned` snaps to calendar boundaries measured from the epoch, which
 * is what makes "the fifth second of every minute" reproducible across
 * restarts and reconnects: interval 60s with offset 5s lands on :05 forever.
 */
export function nextPollAt(
  now: number,
  mode: ScheduleMode,
  intervalMs: number,
  offsetMs: number,
): number {
  if (mode === "interval") return now + intervalMs;

  const shifted = now - offsetMs;
  const remainder = ((shifted % intervalMs) + intervalMs) % intervalMs;
  let next = shifted - remainder + intervalMs + offsetMs;
  if (next <= now) next += intervalMs;
  return next;
}

/* --------------------------------------------------------------- clock skew */

export type ClockState = "unknown" | "ok" | "warn" | "critical";

export interface ClockThresholds {
  warnMs: number;
  criticalMs: number;
}

export interface SkewSample {
  /** Reference instant the sample belongs to — request time plus half the RTT. */
  at: number;
  skewMs: number;
  roundTripMs: number;
}

export interface ClockReport {
  state: ClockState;
  skewMs: number;
  medianMs: number;
  minMs: number;
  maxMs: number;
  driftPerDayMs: number;
  driftDetermined: boolean;
  fit: number;
  samples: number;
  roundTripMs: number;
}

/** Minimum samples before a drift line means anything. */
export const MIN_DRIFT_SAMPLES = 3;

export function classifySkew(skewMs: number, thresholds: ClockThresholds): ClockState {
  const magnitude = Math.abs(skewMs);
  if (magnitude >= thresholds.criticalMs) return "critical";
  if (magnitude >= thresholds.warnMs) return "warn";
  return "ok";
}

/**
 * Skew of one exchange.
 *
 * The device sampled its clock somewhere inside the round trip; assuming the
 * midpoint is the standard compromise and removes the line latency that would
 * otherwise be indistinguishable from a slow clock.
 */
export function sampleSkew(requestedAt: number, receivedAt: number, deviceTime: number): SkewSample {
  const roundTripMs = Math.max(0, receivedAt - requestedAt);
  const reference = requestedAt + roundTripMs / 2;
  return { at: reference, skewMs: deviceTime - reference, roundTripMs };
}

function median(values: number[]): number {
  if (values.length === 0) return 0;
  const sorted = [...values].sort((a, b) => a - b);
  const middle = Math.floor(sorted.length / 2);
  return sorted.length % 2 === 1
    ? sorted[middle]!
    : (sorted[middle - 1]! + sorted[middle]!) / 2;
}

/**
 * Drift rate in ms/day by least squares over the sample window, with the
 * coefficient of determination so the UI can say how well the line actually
 * fits — a confident-looking drift figure derived from noise is worse than
 * no figure at all.
 */
export function computeDrift(samples: SkewSample[]): { perDayMs: number; determined: boolean; fit: number } {
  if (samples.length < MIN_DRIFT_SAMPLES) return { perDayMs: 0, determined: false, fit: 0 };

  const origin = samples[0]!.at;
  const xs = samples.map((sample) => (sample.at - origin) / 1000);
  const ys = samples.map((sample) => sample.skewMs / 1000);
  const meanX = xs.reduce((total, value) => total + value, 0) / xs.length;
  const meanY = ys.reduce((total, value) => total + value, 0) / ys.length;

  let covariance = 0;
  let varianceX = 0;
  for (let index = 0; index < xs.length; index += 1) {
    const dx = xs[index]! - meanX;
    covariance += dx * (ys[index]! - meanY);
    varianceX += dx * dx;
  }
  if (varianceX === 0) return { perDayMs: 0, determined: false, fit: 0 };

  const slope = covariance / varianceX;
  const intercept = meanY - slope * meanX;

  let residual = 0;
  let totalVariance = 0;
  for (let index = 0; index < xs.length; index += 1) {
    const predicted = intercept + slope * xs[index]!;
    residual += (ys[index]! - predicted) ** 2;
    totalVariance += (ys[index]! - meanY) ** 2;
  }
  const fit = totalVariance === 0 ? 0 : 1 - residual / totalVariance;

  return { perDayMs: slope * 86400 * 1000, determined: true, fit: Number.isFinite(fit) ? fit : 0 };
}

/**
 * Folds a sample window into the report the UI renders.
 *
 * State is classified on the *median* once there are enough samples, not on
 * the latest one: a single late response should not raise an alarm, and on a
 * noisy line it otherwise would, every few minutes.
 */
export function buildClockReport(
  samples: SkewSample[],
  thresholds: ClockThresholds,
): ClockReport {
  if (samples.length === 0) {
    return {
      state: "unknown",
      skewMs: 0,
      medianMs: 0,
      minMs: 0,
      maxMs: 0,
      driftPerDayMs: 0,
      driftDetermined: false,
      fit: 0,
      samples: 0,
      roundTripMs: 0,
    };
  }

  const skews = samples.map((sample) => sample.skewMs);
  const last = samples[samples.length - 1]!;
  const medianMs = median(skews);
  const drift = computeDrift(samples);
  const classified = samples.length >= MIN_DRIFT_SAMPLES ? medianMs : last.skewMs;

  return {
    state: classifySkew(classified, thresholds),
    skewMs: last.skewMs,
    medianMs,
    minMs: Math.min(...skews),
    maxMs: Math.max(...skews),
    driftPerDayMs: drift.perDayMs,
    driftDetermined: drift.determined,
    fit: drift.fit,
    samples: samples.length,
    roundTripMs: last.roundTripMs,
  };
}

/* ------------------------------------------------------------ device health */

export type HealthState = "unknown" | "online" | "degraded" | "offline";

export interface HealthPolicy {
  degradeAfter: number;
  offlineAfter: number;
  recoverAfter: number;
}

export interface HealthSnapshot {
  state: HealthState;
  consecutiveFailures: number;
  consecutiveSuccesses: number;
  successes: number;
  failures: number;
  /** Rolling window of outcomes, newest last. */
  window: boolean[];
  latencies: number[];
}

export const EMPTY_HEALTH: HealthSnapshot = {
  state: "unknown",
  consecutiveFailures: 0,
  consecutiveSuccesses: 0,
  successes: 0,
  failures: 0,
  window: [],
  latencies: [],
};

const WINDOW_SIZE = 60;

/**
 * State machine with hysteresis.
 *
 * Thresholds in both directions are what stop a flapping line from producing a
 * stream of alarms: one bad poll is not an outage, and one good poll after an
 * outage is not a recovery.
 */
export function recordOutcome(
  snapshot: HealthSnapshot,
  success: boolean,
  policy: HealthPolicy,
  latencyMs = 0,
): HealthSnapshot {
  const window = [...snapshot.window, success].slice(-WINDOW_SIZE);
  const latencies = success ? [...snapshot.latencies, latencyMs].slice(-WINDOW_SIZE) : snapshot.latencies;

  const consecutiveFailures = success ? 0 : snapshot.consecutiveFailures + 1;
  const consecutiveSuccesses = success ? snapshot.consecutiveSuccesses + 1 : 0;

  let state = snapshot.state;
  if (success) {
    if (state === "online" || state === "unknown") state = "online";
    else if (consecutiveSuccesses >= policy.recoverAfter) state = "online";
  } else if (consecutiveFailures >= policy.offlineAfter) state = "offline";
  else if (consecutiveFailures >= policy.degradeAfter) state = "degraded";

  return {
    state,
    consecutiveFailures,
    consecutiveSuccesses,
    successes: snapshot.successes + (success ? 1 : 0),
    failures: snapshot.failures + (success ? 0 : 1),
    window,
    latencies,
  };
}

export function availability(snapshot: HealthSnapshot): number {
  if (snapshot.window.length === 0) return 0;
  const ok = snapshot.window.filter(Boolean).length;
  return ok / snapshot.window.length;
}

export function percentile(values: number[], quantile: number): number {
  if (values.length === 0) return 0;
  const sorted = [...values].sort((a, b) => a - b);
  const index = Math.min(sorted.length - 1, Math.max(0, Math.round((sorted.length - 1) * quantile)));
  return sorted[index]!;
}
