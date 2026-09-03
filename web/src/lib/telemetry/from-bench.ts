import { availability, percentile, type ClockReport } from "@/lib/bench/domain";
import type { BenchState } from "@/stores/bench-store";

import type {
  CounterView,
  FaultView,
  HealthView,
  LimitsView,
  ScheduleView,
  Telemetry,
} from "./types";

/**
 * The bench engine, in the console's vocabulary.
 *
 * The engine already speaks most of it — the log rows and the schedule are
 * shared types — so this is mostly a projection. The two places it does real
 * work are the ones where the live source has an answer the bench keeps in a
 * different form: availability and the latency percentiles, which the gateway
 * reports directly and the bench has to fold out of its outcome window.
 *
 * Doing that fold here rather than in the dashboard is the point. It used to
 * live in the panel, which meant the panel could only ever render the bench.
 */

export function benchHealth(state: BenchState): HealthView {
  const { health } = state;
  return {
    state: health.state,
    availability: availability(health),
    window: health.window,
    windowSamples: health.window.length,
    consecutiveFailures: health.consecutiveFailures,
    consecutiveSuccesses: health.consecutiveSuccesses,
    latencyP50Ms: percentile(health.latencies, 0.5),
    latencyP95Ms: percentile(health.latencies, 0.95),
  };
}

export function benchCounters(state: BenchState): CounterView {
  const { counters } = state;
  return {
    successfulReads: counters.successfulReads,
    failedReads: counters.failedReads,
    retries: counters.retries,
    protocolErrors: counters.protocolErrors,
    reconnects: counters.reconnects,
    connections: counters.connections,
  };
}

export function benchSchedule(state: BenchState): ScheduleView {
  return {
    mode: state.gateway.scheduleMode,
    intervalMs: state.gateway.intervalMs,
    offsetMs: state.gateway.offsetMs,
    nextPollAt: state.nextPollAt,
  };
}

/**
 * Schedule ticks: the request that *opened* each cycle.
 *
 * Retries are requests too, but they are not schedule events, and including
 * them would make a perfectly regular schedule look ragged on the timeline.
 * The bench knows which is which because it numbered them.
 */
export function benchTicks(state: BenchState): number[] {
  return state.events
    .filter(
      (event) =>
        event.direction === "tx" && event.attempt === 0 && event.command === "read-time",
    )
    .map((event) => event.at);
}

export function benchLimits(state: BenchState): LimitsView {
  return {
    requestTimeoutMs: state.gateway.requestTimeoutMs,
    // The bench dials a socket that does not exist, so a connect timeout has
    // nothing to time out; the gateway's own value is reported as zero.
    connectTimeoutMs: 0,
    retryAttempts: state.gateway.retryAttempts,
  };
}

export function benchFaults(state: BenchState): FaultView {
  const { faults } = state;
  return {
    responseDelayMs: faults.responseDelayMs,
    badChecksumProb: faults.badChecksumProb,
    fragmentProb: faults.fragmentProb,
    fragmentDelayMs: faults.fragmentDelayMs,
    noResponse: faults.noResponse,
    closeAfterRequest: faults.closeAfterRequest,
    clockOffsetMs: faults.clockOffsetMs,
    clockDriftPerDayMs: faults.clockDriftPerDayMs,
  };
}

/**
 * The clock report is passed in rather than computed here: it is derived from
 * the sample window on every call, and the hook that owns this projection
 * memoises it so a re-render does not produce a new object identity for a
 * value that has not changed.
 */
export function benchTelemetry(state: BenchState, clock: ClockReport): Telemetry {
  return {
    source: "bench",
    // The bench runs in this tab; there is nothing it could fail to reach.
    link: "ready",
    error: null,
    editable: true,
    running: state.running,
    connected: state.connected,
    checksumMode: state.checksumMode,
    target: null,
    events: state.events,
    ticks: benchTicks(state),
    clock,
    thresholds: state.gateway.thresholds,
    samples: state.samples,
    health: benchHealth(state),
    policy: state.gateway.policy,
    counters: benchCounters(state),
    schedule: benchSchedule(state),
    limits: benchLimits(state),
    identity: state.identityRead,
    lastDeviceTime: state.lastDeviceTime,
    lastCycleAt: state.lastCycleAt,
    // A simulated device is a device, not a fleet.
    fleet: null,
    faults: benchFaults(state),
  };
}
