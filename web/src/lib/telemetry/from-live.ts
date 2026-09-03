import type { ApiEvent, ApiFleet, ApiGatewayStatus, ApiHistory } from "@/lib/api/schema";
import type {
  ClockReport,
  ClockState,
  HealthState,
  ScheduleMode,
  SkewSample,
} from "@/lib/bench/domain";
import { CHECKSUM_MODES, parseHex, type ChecksumMode } from "@/lib/ft12";

import type {
  BenchErrorCode,
  CounterView,
  EventDirection,
  FleetView,
  HealthView,
  IdentityView,
  LogEvent,
  ScheduleView,
} from "./types";

/**
 * The Go gateway's answers, translated into the console's vocabulary.
 *
 * Everything here is pure: a payload in, a view out. That is what makes the
 * live source testable without a server — the fixtures are the JSON the API
 * actually returns, and the assertions are about what an operator would read
 * off the screen.
 *
 * The translation is deliberately conservative. Where the gateway knows
 * something the bench does not (the drift fit over its own longer window), its
 * figure is used as given rather than recomputed from the truncated history it
 * sends; where the gateway does not know something the bench does (which retry
 * a frame belonged to), the field reads zero rather than being guessed.
 */

/* ------------------------------------------------------------------- times */

/** RFC3339 to epoch ms; 0 for absent or unparseable, which reads as "never". */
export function instantToMs(value: string | undefined): number {
  if (!value) return 0;
  const parsed = Date.parse(value);
  return Number.isNaN(parsed) ? 0 : parsed;
}

/* ------------------------------------------------------------------ states */

const HEALTH_STATES: readonly HealthState[] = ["unknown", "online", "degraded", "offline"];
const CLOCK_STATES: readonly ClockState[] = ["unknown", "ok", "warn", "critical"];

/**
 * A state name from the wire, narrowed to one this console can colour.
 *
 * The gateway sends its state as a string, and an unrecognised one has to land
 * somewhere: `unknown` is the honest place, because it is exactly what the
 * console knows about it.
 */
export function toHealthState(value: string): HealthState {
  return HEALTH_STATES.includes(value as HealthState) ? (value as HealthState) : "unknown";
}

export function toClockState(value: string): ClockState {
  return CLOCK_STATES.includes(value as ClockState) ? (value as ClockState) : "unknown";
}

export function toChecksumMode(value: string): ChecksumMode {
  return CHECKSUM_MODES.includes(value as ChecksumMode) ? (value as ChecksumMode) : "sum";
}

export function toScheduleMode(value: string): ScheduleMode {
  return value === "aligned" ? "aligned" : "interval";
}

/* ------------------------------------------------------------------ events */

const DIRECTIONS: Record<string, EventDirection> = {
  TX: "tx",
  RX: "rx",
  ERR: "err",
  ERROR: "err",
  SYSTEM: "sys",
};

/**
 * Faults the console has its own words for, recognised in the gateway's.
 *
 * These are matched on the sentinel wording the Go errors are built from, not
 * on an error code — the control plane has never had one. The list is short on
 * purpose: anything not on it keeps the gateway's own message rather than
 * being forced into the nearest category, because a mislabelled fault is
 * harder to debug than an untranslated one.
 */
const ERROR_PATTERNS: readonly (readonly [RegExp, BenchErrorCode])[] = [
  [/invalid checksum/i, "invalidChecksum"],
  [/no response frame/i, "noResponse"],
  [/deadline exceeded|timeout|timed out/i, "timeout"],
  [/EOF|connection reset|connection aborted|use of closed/i, "connectionClosed"],
];

export function classifyError(message: string): BenchErrorCode | undefined {
  for (const [pattern, code] of ERROR_PATTERNS) {
    if (pattern.test(message)) return code;
  }
  return undefined;
}

/**
 * A gateway frame record as a log row.
 *
 * The lane a row is drawn in comes from its direction rather than from the
 * `service` field: the gateway records both halves of the exchange from its
 * own side, so an RX row is the device speaking even though the gateway wrote
 * it down.
 */
export function toLogEvent(event: ApiEvent): LogEvent {
  const direction = DIRECTIONS[event.direction.toUpperCase()] ?? "sys";
  const message = event.error || event.message;

  return {
    id: event.id,
    at: instantToMs(event.timestamp),
    direction,
    source: direction === "rx" ? "device" : "gateway",
    command: event.command,
    bytes: event.rawHex ? parseHex(event.rawHex).bytes : [],
    errorCode: direction === "err" ? classifyError(message) : undefined,
    detail: message || undefined,
    // The gateway's history records neither the poll cycle nor the retry
    // attempt a frame belonged to, so these read zero rather than being
    // inferred from adjacent rows.
    cycle: 0,
    attempt: 0,
  };
}

/** Oldest first, which is the order every view in this console reads in. */
export function toLogEvents(events: ApiEvent[]): LogEvent[] {
  return events
    .map(toLogEvent)
    .sort((left, right) => left.at - right.at || left.id - right.id);
}

/* ---------------------------------------------------------------- readouts */

export function toClockReport(status: ApiGatewayStatus): ClockReport {
  const clock = status.clock;
  return {
    state: toClockState(clock.state),
    skewMs: clock.skewMs,
    medianMs: clock.medianSkewMs,
    minMs: clock.minSkewMs,
    maxMs: clock.maxSkewMs,
    driftPerDayMs: clock.driftPerDayMs,
    driftDetermined: clock.driftDetermined,
    fit: clock.driftFit,
    samples: clock.samples,
    roundTripMs: clock.roundTripMs,
  };
}

export function toSkewSamples(history: ApiHistory): SkewSample[] {
  return history.clockSamples.map((sample) => ({
    at: instantToMs(sample.at),
    skewMs: sample.skewMs,
    roundTripMs: sample.roundTripMs,
  }));
}

export function toHealthView(status: ApiGatewayStatus, history: ApiHistory): HealthView {
  return {
    state: toHealthState(status.health.state),
    availability: status.health.availability,
    window: history.healthOutcomes.map((outcome) => outcome.success),
    windowSamples: status.health.windowSamples,
    consecutiveFailures: status.health.consecutiveFailures,
    consecutiveSuccesses: status.health.consecutiveSuccesses,
    latencyP50Ms: status.health.latencyP50Ms,
    latencyP95Ms: status.health.latencyP95Ms,
  };
}

/**
 * The counters, mapped onto the six the console shows.
 *
 * `connections` is the gateway's connection *attempts* rather than a separate
 * session count: the gateway does not keep one, and attempts is the number
 * that answers the same question — how many times has this link been set up.
 */
export function toCounters(status: ApiGatewayStatus): CounterView {
  return {
    successfulReads: status.successfulReads,
    failedReads: status.failedReads,
    retries: status.retry.totalRetries,
    protocolErrors: status.protocolErrors,
    reconnects: status.reconnects,
    connections: status.connectionAttempts,
  };
}

export function toSchedule(status: ApiGatewayStatus): ScheduleView {
  return {
    mode: toScheduleMode(status.schedule.mode),
    intervalMs: status.schedule.intervalMs || status.pollingIntervalMs,
    offsetMs: status.schedule.offsetMs,
    nextPollAt: instantToMs(status.schedule.nextPollAt),
  };
}

export function toIdentity(status: ApiGatewayStatus): IdentityView | null {
  if (!status.identity.known) return null;
  return {
    model: status.identity.model,
    serial: status.identity.serial,
    firmware: status.identity.firmware,
  };
}

/**
 * Schedule ticks, as observed on the wire.
 *
 * The bench can tell a first attempt from a retry and plots only the former.
 * The gateway cannot, so every read-time request appears here — a retried poll
 * shows as a cluster rather than as a single tick, which is a truthful
 * rendering of what actually went out on the line.
 */
export function toTicks(events: LogEvent[]): number[] {
  return events
    .filter((event) => event.direction === "tx" && event.command === "read-time")
    .map((event) => event.at);
}

export function toFleet(fleet: ApiFleet): FleetView {
  return {
    devices: fleet.devices.map((device) => ({
      id: device.deviceId,
      name: device.deviceName || device.deviceId,
      target: device.targetAddr,
      running: device.state === "running" || device.state === "degraded",
      health: toHealthState(device.health.state),
      clock: toClockState(device.clock.state),
      skewMs: device.clock.skewMs,
      availability: device.health.availability,
      samples: device.clock.samples,
    })),
    online: fleet.summary.online,
    degraded: fleet.summary.degraded,
    offline: fleet.summary.offline,
    clockWarn: fleet.summary.clockWarn,
    clockCritical: fleet.summary.clockCritical,
    worstSkewMs: fleet.summary.worstClockSkewMs,
    worstDeviceId: fleet.summary.worstClockDeviceId,
  };
}
