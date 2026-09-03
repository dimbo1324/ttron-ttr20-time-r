"use client";

import { create } from "zustand";

import { api, ApiError } from "@/lib/api/client";
import type {
  ApiFaultMode,
  ApiFleet,
  ApiGatewaySettings,
  ApiGatewayStatus,
  ApiHistory,
} from "@/lib/api/schema";
import type { ClockReport } from "@/lib/bench/domain";
import {
  toChecksumMode,
  toClockReport,
  toCounters,
  toFleet,
  toHealthView,
  toIdentity,
  toLogEvents,
  toSchedule,
  toSkewSamples,
  toTicks,
} from "@/lib/telemetry/from-live";
import type {
  CounterView,
  FaultView,
  HealthView,
  LinkState,
  LogEvent,
  Telemetry,
} from "@/lib/telemetry/types";

/**
 * The live source: the console pointed at a running Go stack.
 *
 * It is a poller rather than a stream because the API is request/response and
 * a bench does not need sub-second fidelity — the gateway's own poll interval
 * is measured in seconds. One refresh fetches the four things a console needs
 * (status, log, history, fleet) in parallel and commits them together, so no
 * panel can ever show a clock from one instant next to counters from another.
 *
 * ## Why a partial failure is not a failure
 *
 * `gateway/fleet` and `gateway/history` are newer than the rest of the API. A
 * gateway that predates them answers 404, and a console that treated that as
 * an outage would go dark against a perfectly healthy stack. So the status
 * call decides whether the link is up, and the other three degrade to what
 * they last held.
 */

const EVENT_LIMIT = 200;

/** How much of the log is kept, matching the bench engine's own ceiling. */
const MAX_EVENTS = 400;

interface LiveState {
  link: LinkState;
  error: string | null;
  status: ApiGatewayStatus | null;
  history: ApiHistory | null;
  fleet: ApiFleet | null;
  events: LogEvent[];
  /** Emulator faults as last read back from the API; null until first read. */
  faults: ApiFaultMode | null;
  /**
   * Why the last settings write was refused, kept apart from `error`.
   *
   * `error` is rewritten by every refresh, a second apart, so a rejection put
   * there would flash and vanish before it could be read. This one survives
   * until the next write, which is how long the reason stays relevant.
   */
  settingsError: string | null;
  /** True while a command (start, stop, fault change) is in flight. */
  busy: boolean;

  refresh: () => Promise<void>;
  start: () => Promise<void>;
  stop: () => Promise<void>;
  loadFaults: () => Promise<void>;
  setFaults: (patch: Partial<ApiFaultMode>) => Promise<void>;
  setSettings: (settings: ApiGatewaySettings) => Promise<void>;
  reset: () => void;
}

const EMPTY: Pick<
  LiveState,
  "link" | "error" | "status" | "history" | "fleet" | "events" | "faults" | "settingsError" | "busy"
> = {
  link: "connecting",
  error: null,
  status: null,
  history: null,
  fleet: null,
  events: [],
  faults: null,
  settingsError: null,
  busy: false,
};

/** Describes a failure in the terms the status strip needs. */
function describe(cause: unknown): { link: LinkState; error: string } {
  if (cause instanceof ApiError) {
    return { link: cause.offline ? "unreachable" : "ready", error: cause.message };
  }
  return { link: "unreachable", error: cause instanceof Error ? cause.message : "unknown error" };
}

export const useLiveStore = create<LiveState>()((set, get) => ({
  ...EMPTY,

  refresh: async () => {
    // Only the status call decides whether the link is up; the other three are
    // allowed to be missing on an older gateway without taking the page down.
    const [status, events, history, fleet] = await Promise.all([
      api.gatewayStatus().then(
        (value) => ({ ok: true as const, value }),
        (cause: unknown) => ({ ok: false as const, cause }),
      ),
      api.gatewayEvents(EVENT_LIMIT).catch(() => null),
      api.gatewayHistory().catch(() => null),
      api.gatewayFleet().catch(() => null),
    ]);

    if (!status.ok) {
      // Everything that describes "now" goes with the status. A stale gauge
      // beside an "API unreachable" notice reads as a current value, which is
      // the one mistake a monitoring view must not make.
      //
      // The frame log stays. Its rows carry their own timestamps, so they
      // cannot be mistaken for the present, and throwing away the evidence an
      // operator is halfway through reading because the API blinked would be
      // gratuitous.
      set({ ...describe(status.cause), status: null, history: null, fleet: null });
      return;
    }

    set((state) => ({
      link: "ready",
      error: status.value.lastError || null,
      status: status.value,
      history: history ?? state.history,
      fleet: fleet ?? state.fleet,
      events: events ? toLogEvents(events).slice(-MAX_EVENTS) : state.events,
    }));
  },

  start: async () => {
    set({ busy: true });
    try {
      set({ status: await api.startPolling(), link: "ready", error: null });
    } catch (cause) {
      set(describe(cause));
    } finally {
      set({ busy: false });
    }
  },

  stop: async () => {
    set({ busy: true });
    try {
      set({ status: await api.stopPolling(), link: "ready", error: null });
    } catch (cause) {
      set(describe(cause));
    } finally {
      set({ busy: false });
    }
  },

  loadFaults: async () => {
    try {
      set({ faults: await api.faultMode() });
    } catch {
      // The emulator is a separate process from the gateway and may well not
      // be running; that is not a reason to report the gateway as down.
    }
  },

  setFaults: async (patch) => {
    const current = get().faults;
    if (!current) return;

    const next = { ...current, ...patch };
    // Shown immediately, then replaced by whatever the emulator accepted --
    // a slider that waits for a round trip before moving feels broken.
    set({ faults: next, busy: true });
    try {
      set({ faults: await api.setFaultMode(next) });
    } catch (cause) {
      set({ faults: current, ...describe(cause) });
    } finally {
      set({ busy: false });
    }
  },

  /**
   * Reconfigures the running gateway.
   *
   * Deliberately not optimistic, unlike the fault sliders. A fault takes
   * effect on the next frame and is trivially reversible; a schedule the
   * gateway rejected must never appear on screen as though it had been
   * accepted, because the next reading under it would look like a fault.
   */
  setSettings: async (settings) => {
    set({ busy: true, settingsError: null });
    try {
      const applied = await api.updateSettings(settings);
      set({ status: applied.status, link: "ready", error: null });
    } catch (cause) {
      const described = describe(cause);
      // A refusal and an outage are different failures and get different
      // places to say so. The gateway refusing a value leaves the link fine,
      // so it goes to the settings banner and not to `error`, which describes
      // the link. An unreachable API sets `error`, and the notice that takes
      // over the page explains it better than a second banner would.
      set(
        described.link === "ready"
          ? { settingsError: described.error }
          : { ...described, settingsError: null },
      );
    } finally {
      set({ busy: false });
    }
  },

  reset: () => set({ ...EMPTY }),
}));

/**
 * The live source as telemetry.
 *
 * Before the first successful refresh there is no status, and the console
 * still has to render something: an empty reading in every field, with `link`
 * carrying the reason. That is why this returns a full `Telemetry` rather than
 * `Telemetry | null` — a nullable source would put a "is it loaded yet" branch
 * into every panel.
 */
export function liveTelemetry(state: LiveState): Telemetry {
  const { status, history, fleet } = state;

  if (!status) {
    return {
      source: "live",
      link: state.link,
      error: state.error,
      editable: false,
      deviceName: null,
      running: false,
      connected: false,
      checksumMode: "sum",
      target: null,
      events: state.events,
      ticks: [],
      clock: EMPTY_CLOCK,
      thresholds: { warnMs: 0, criticalMs: 0 },
      samples: [],
      health: EMPTY_HEALTH_VIEW,
      policy: { degradeAfter: 0, offlineAfter: 0, recoverAfter: 0 },
      counters: EMPTY_COUNTERS,
      schedule: { mode: "interval", intervalMs: 0, offsetMs: 0, nextPollAt: 0 },
      limits: { requestTimeoutMs: 0, connectTimeoutMs: 0, retryAttempts: 0, retryDelayMs: 0 },
      identity: null,
      lastDeviceTime: null,
      // No status means the link is down or has never been up, and refresh
      // drops the fleet along with it -- a fleet table beside an empty
      // reading would be a stale table presented as current.
      fleet: null,
      lastCycleAt: 0,
      faults: toFaults(state.faults),
      settingsError: state.settingsError,
    };
  }

  // A gateway too old to answer /history still reports its aggregates; the
  // charts simply have nothing to draw until it does.
  const windows: ApiHistory = history ?? { clockSamples: [], healthOutcomes: [] };

  return {
    source: "live",
    link: state.link,
    error: state.error,
    // Settings can be written, but only while the link is up: a control that
    // cannot reach the gateway is shown as a value rather than as a knob.
    editable: state.link === "ready",
    deviceName: status.deviceName || status.deviceId || null,
    running: status.state === "running" || status.state === "degraded",
    connected: status.connected,
    checksumMode: toChecksumMode(status.checksumMode),
    target: status.targetAddr || null,
    events: state.events,
    ticks: toTicks(state.events),
    clock: toClockReport(status),
    thresholds: {
      warnMs: status.clock.warnThresholdMs,
      criticalMs: status.clock.criticalThresholdMs,
    },
    samples: toSkewSamples(windows),
    health: toHealthView(status, windows),
    policy: {
      degradeAfter: status.health.degradeAfter,
      offlineAfter: status.health.offlineAfter,
      recoverAfter: status.health.recoverAfter,
    },
    counters: toCounters(status),
    schedule: toSchedule(status),
    limits: {
      requestTimeoutMs: status.requestTimeoutMs,
      connectTimeoutMs: status.connectTimeoutMs,
      retryAttempts: status.retry.attempts,
      retryDelayMs: status.retry.delayMs,
    },
    identity: toIdentity(status),
    lastDeviceTime: status.lastDeviceTime ? Date.parse(status.lastDeviceTime) : null,
    lastCycleAt: status.lastSuccessfulReadTime ? Date.parse(status.lastSuccessfulReadTime) : 0,
    fleet: fleet ? toFleet(fleet) : null,
    faults: toFaults(state.faults),
    settingsError: state.settingsError,
  };
}

/**
 * The emulator's fault mode as the console describes faults.
 *
 * The two clock fields come back null: the Go emulator answers with the
 * device's real clock and has no offset to inject. Null rather than zero,
 * because zero would claim the clock had been checked and found correct.
 */
function toFaults(faults: ApiFaultMode | null): FaultView | null {
  if (!faults) return null;
  return {
    responseDelayMs: faults.responseDelayMs,
    // The emulator carries a flag and a probability, and raises the flag
    // whenever the probability is above zero. So the flag alone says nothing;
    // only a flag with *no* probability means "every frame", which is the one
    // case that reads back as certainty. Treating the flag as certainty on its
    // own would snap a 35% slider to 100% on the first refresh.
    badChecksumProb:
      faults.corruptChecksum && faults.corruptChecksumProbability === 0
        ? 1
        : faults.corruptChecksumProbability,
    fragmentProb:
      faults.fragmentResponse && faults.fragmentProbability === 0
        ? 1
        : faults.fragmentProbability,
    fragmentDelayMs: faults.fragmentDelayMs,
    noResponse: faults.noResponse,
    closeAfterRequest: faults.closeAfterRequest,
    clockOffsetMs: null,
    clockDriftPerDayMs: null,
  };
}

const EMPTY_CLOCK: ClockReport = {
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

const EMPTY_HEALTH_VIEW: HealthView = {
  state: "unknown",
  availability: 0,
  window: [],
  windowSamples: 0,
  consecutiveFailures: 0,
  consecutiveSuccesses: 0,
  latencyP50Ms: 0,
  latencyP95Ms: 0,
};

const EMPTY_COUNTERS: CounterView = {
  successfulReads: 0,
  failedReads: 0,
  retries: 0,
  protocolErrors: 0,
  reconnects: 0,
  connections: 0,
};
