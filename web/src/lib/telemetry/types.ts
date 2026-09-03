import type {
  ClockReport,
  ClockState,
  ClockThresholds,
  HealthPolicy,
  HealthState,
  ScheduleMode,
  SkewSample,
} from "@/lib/bench/domain";
import type { ChecksumMode } from "@/lib/ft12";

/**
 * The vocabulary every readout in this console is written against.
 *
 * The bench engine and the real Go gateway measure the same things and say
 * them differently — one holds a `HealthSnapshot` with a raw outcome window,
 * the other answers HTTP with an availability ratio and a percentile table.
 * Left as-is, every panel would have to know which one it was looking at, and
 * "which source am I on" would end up spelled out in a dozen components.
 *
 * So both are normalised into `Telemetry` and the panels read only this. The
 * dashboard does not know whether the number it is drawing came from a
 * simulation in this tab or from a meter on a bench, and it should not: the
 * whole point of the bench is that it behaves like the real thing.
 */

export type TelemetrySource = "bench" | "live";

/**
 * How the console stands with respect to its source.
 *
 * The bench is always `ready` — it cannot fail to be reachable. The other
 * three exist for the live source, and `unreachable` is the common one: the
 * Go API is simply not running, which needs different wording from a gateway
 * that answered with an error.
 */
export type LinkState = "ready" | "connecting" | "unreachable";

export type EventDirection = "tx" | "rx" | "err" | "sys";

/** Faults the console can name in both languages. */
export type BenchErrorCode = "invalidChecksum" | "noResponse" | "timeout" | "connectionClosed";

/** System notes the console can name in both languages. */
export type BenchNote =
  | "pollingStarted"
  | "pollingStopped"
  | "deviceStateChanged"
  | "clockStateChanged"
  | "reconnected";

/**
 * One row of the exchange log.
 *
 * Both sources produce these. Where they differ is in what they can fill in:
 * the bench knows the cycle and attempt of every frame because it generated
 * them, while the gateway's history records neither, so those read zero for a
 * live event rather than being invented.
 */
export interface LogEvent {
  id: number;
  /** Wall clock of the event, ms since epoch. */
  at: number;
  direction: EventDirection;
  source: "gateway" | "device";
  command: string;
  bytes: number[];
  /** Dictionary key for a system note, when the event is not a frame. */
  note?: BenchNote;
  noteArgs?: { from: string; to: string };
  errorCode?: BenchErrorCode;
  /**
   * Free text from the source, for anything the console has no name for.
   * The live gateway can report faults this UI has never heard of, and
   * showing its own words beats showing nothing.
   */
  detail?: string;
  cycle: number;
  attempt: number;
  latencyMs?: number;
}

/**
 * Reachability, as a reader needs it.
 *
 * `window` is the raw run of outcomes, oldest first — the strip that shows
 * *which* polls failed rather than only how many. Percentiles arrive
 * pre-computed because the live source computes them over a longer window
 * than it sends, and recomputing from a truncated copy would quietly disagree
 * with the gateway's own figure.
 */
export interface HealthView {
  state: HealthState;
  availability: number;
  window: boolean[];
  windowSamples: number;
  consecutiveFailures: number;
  consecutiveSuccesses: number;
  latencyP50Ms: number;
  latencyP95Ms: number;
}

export interface CounterView {
  successfulReads: number;
  failedReads: number;
  retries: number;
  protocolErrors: number;
  reconnects: number;
  connections: number;
}

export interface ScheduleView {
  mode: ScheduleMode;
  intervalMs: number;
  offsetMs: number;
  /** Next poll instant, ms since epoch; 0 when nothing is scheduled. */
  nextPollAt: number;
}

/**
 * How long one exchange is given, and how many tries it gets.
 *
 * Separate from the schedule because they answer a different question: the
 * schedule says when a poll starts, these say when it is given up on.
 */
export interface LimitsView {
  requestTimeoutMs: number;
  connectTimeoutMs: number;
  retryAttempts: number;
}

export interface IdentityView {
  model: string;
  serial: string;
  firmware: string;
}

/** One device as the fleet table lists it. */
export interface FleetDevice {
  id: string;
  name: string;
  target: string;
  running: boolean;
  health: HealthState;
  clock: ClockState;
  skewMs: number;
  availability: number;
  samples: number;
}

export interface FleetView {
  devices: FleetDevice[];
  online: number;
  degraded: number;
  offline: number;
  clockWarn: number;
  clockCritical: number;
  worstSkewMs: number;
  worstDeviceId: string;
}

/**
 * The faults the source is currently injecting.
 *
 * Both sources can produce the six transport faults; only the bench can move
 * the device clock, because the Go emulator has no such setting. Those two
 * fields are therefore nullable rather than zero — zero is a real value that
 * means "the clock is correct", and an operator must be able to tell it apart
 * from "this source cannot do that at all".
 */
export interface FaultView {
  responseDelayMs: number;
  badChecksumProb: number;
  fragmentProb: number;
  fragmentDelayMs: number;
  noResponse: boolean;
  closeAfterRequest: boolean;
  clockOffsetMs: number | null;
  clockDriftPerDayMs: number | null;
}

/** How many faults are switched on, for the "device is not healthy" badge. */
export function activeFaultCount(faults: FaultView | null): number {
  if (!faults) return 0;
  return [
    faults.responseDelayMs > 0,
    faults.badChecksumProb > 0,
    faults.fragmentProb > 0,
    faults.noResponse,
    faults.closeAfterRequest,
    faults.clockOffsetMs !== null && faults.clockOffsetMs !== 0,
    faults.clockDriftPerDayMs !== null && faults.clockDriftPerDayMs !== 0,
  ].filter(Boolean).length;
}

export interface Telemetry {
  source: TelemetrySource;
  link: LinkState;
  /** What went wrong, in the source's own words; null when nothing did. */
  error: string | null;
  /**
   * Whether the settings panels may write.
   *
   * False on the live source: interval, thresholds and the health policy are
   * the gateway process's own configuration, and the control plane has no
   * setter for them. Showing an editable control that silently does nothing
   * would be worse than showing the real value and saying it is read-only.
   */
  editable: boolean;
  running: boolean;
  connected: boolean;
  checksumMode: ChecksumMode;
  /** The device the gateway is polling, when the source knows of one. */
  target: string | null;
  events: LogEvent[];
  /** Instants of the requests that opened a poll — the schedule, as observed. */
  ticks: number[];
  clock: ClockReport;
  thresholds: ClockThresholds;
  samples: SkewSample[];
  health: HealthView;
  policy: HealthPolicy;
  counters: CounterView;
  schedule: ScheduleView;
  limits: LimitsView;
  identity: IdentityView | null;
  /** Device wall clock as of the last successful read, ms since epoch. */
  lastDeviceTime: number | null;
  /** When the last exchange finished, ms since epoch; 0 when none has. */
  lastCycleAt: number;
  /** Every polled device, when the source has more than its own. */
  fleet: FleetView | null;
  /** Faults in effect; null when the source has no device to inject them into. */
  faults: FaultView | null;
}
