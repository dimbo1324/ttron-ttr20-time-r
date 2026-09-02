"use client";

import { useMemo } from "react";
import { create } from "zustand";

import {
  availability,
  buildClockReport,
  EMPTY_HEALTH,
  nextPollAt,
  recordOutcome,
  sampleSkew,
  type ClockReport,
  type ClockThresholds,
  type HealthPolicy,
  type HealthSnapshot,
  type ScheduleMode,
  type SkewSample,
} from "@/lib/bench/domain";
import {
  buildReadIdentityRequest,
  buildReadIdentityResponse,
  buildReadTimeRequest,
  buildReadTimeResponse,
  encodeFrame,
  formatHex,
  RESPONSE_BIT,
  type ChecksumMode,
} from "@/lib/ft12";

/**
 * The bench engine.
 *
 * This is a working model of the Go stack — device, line and gateway — running
 * in the browser. It exists because the console has to teach the protocol with
 * nothing else installed: a student opens the page and immediately sees real
 * frames, real checksums and a real failure when a fault is switched on.
 *
 * Everything it produces is derived from the same primitives the Go side uses
 * (`@/lib/ft12`), so a frame copied out of the log decodes identically in the
 * analyzer and on the wire.
 *
 * ## Why one exchange resolves at one instant
 *
 * A poll — request, retries, answer — is computed in a single pass at the poll
 * moment, with each event carrying its own simulated timestamp, rather than
 * being spread across real timers. The timeline the operator sees is identical
 * either way, and this version has no partially-completed exchange that a
 * paused tab, a fault toggled mid-flight or a React re-render could tear.
 */

export type EventDirection = "tx" | "rx" | "err" | "sys";

export interface BenchEvent {
  id: number;
  /** Simulated wall clock of the event, ms since epoch. */
  at: number;
  direction: EventDirection;
  source: "gateway" | "device";
  command: string;
  bytes: number[];
  /** Dictionary key for a system note, when the event is not a frame. */
  note?: string;
  noteArgs?: Record<string, string | number>;
  errorCode?: string;
  cycle: number;
  attempt: number;
  latencyMs?: number;
}

export interface DeviceFaults {
  responseDelayMs: number;
  badChecksumProb: number;
  fragmentProb: number;
  fragmentDelayMs: number;
  noResponse: boolean;
  closeAfterRequest: boolean;
  clockOffsetMs: number;
  clockDriftPerDayMs: number;
}

export interface DeviceIdentity {
  model: string;
  serial: string;
  firmware: string;
}

export interface GatewaySettings {
  scheduleMode: ScheduleMode;
  intervalMs: number;
  offsetMs: number;
  requestTimeoutMs: number;
  retryAttempts: number;
  thresholds: ClockThresholds;
  policy: HealthPolicy;
  identityProbe: boolean;
}

export interface BenchCounters {
  successfulReads: number;
  failedReads: number;
  reconnects: number;
  retries: number;
  protocolErrors: number;
  exhaustedPolls: number;
  connections: number;
}

interface BenchState {
  running: boolean;
  connected: boolean;
  checksumMode: ChecksumMode;
  adapterAddress: number;

  faults: DeviceFaults;
  identity: DeviceIdentity;
  gateway: GatewaySettings;

  events: BenchEvent[];
  samples: SkewSample[];
  health: HealthSnapshot;
  counters: BenchCounters;

  cycle: number;
  nextPollAt: number;
  lastCycleAt: number;
  lastDeviceTime: number | null;
  identityRead: DeviceIdentity | null;
  clockOrigin: number;

  setChecksumMode: (mode: ChecksumMode) => void;
  setAdapterAddress: (address: number) => void;
  patchFaults: (patch: Partial<DeviceFaults>) => void;
  setIdentity: (patch: Partial<DeviceIdentity>) => void;
  patchGateway: (patch: Partial<GatewaySettings>) => void;
  start: () => void;
  stop: () => void;
  reset: () => void;
  clearEvents: () => void;
  /** Advances the simulation to `now`; called by the engine ticker. */
  tick: (now: number) => void;
}

const MAX_EVENTS = 400;
const MAX_SAMPLES = 120;

/** Line latency floor and jitter, in ms — a real RS-485/TCP hop is never 0. */
const BASE_LATENCY_MS = 6;
const JITTER_MS = 9;

export const DEFAULT_FAULTS: DeviceFaults = {
  responseDelayMs: 0,
  badChecksumProb: 0,
  fragmentProb: 0,
  fragmentDelayMs: 40,
  noResponse: false,
  closeAfterRequest: false,
  clockOffsetMs: 0,
  clockDriftPerDayMs: 0,
};

export const DEFAULT_GATEWAY: GatewaySettings = {
  scheduleMode: "aligned",
  intervalMs: 5000,
  offsetMs: 0,
  requestTimeoutMs: 1500,
  retryAttempts: 2,
  thresholds: { warnMs: 2000, criticalMs: 30000 },
  policy: { degradeAfter: 3, offlineAfter: 10, recoverAfter: 2 },
  identityProbe: true,
};

let eventId = 0;
const nextEventId = () => {
  eventId += 1;
  return eventId;
};

/** Device wall clock at `now`: real time plus a constant offset and drift. */
function deviceClock(now: number, faults: DeviceFaults, origin: number): number {
  const elapsedDays = (now - origin) / 86_400_000;
  return now + faults.clockOffsetMs + faults.clockDriftPerDayMs * elapsedDays;
}

export const useBenchStore = create<BenchState>()((set, get) => ({
  running: false,
  connected: false,
  checksumMode: "sum",
  adapterAddress: 1,

  faults: { ...DEFAULT_FAULTS },
  identity: { model: "TTR20", serial: "SN-0000042", firmware: "1.2.3" },
  gateway: { ...DEFAULT_GATEWAY },

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
  clockOrigin: Date.now(),

  setChecksumMode: (mode) => set({ checksumMode: mode }),
  setAdapterAddress: (address) => set({ adapterAddress: address & 0xff }),

  patchFaults: (patch) =>
    set((state) => {
      const faults = { ...state.faults, ...patch };
      // Restarting the drift origin on a drift change keeps the accumulated
      // offset continuous: without it, changing the rate would retroactively
      // rewrite how far the clock has already wandered.
      const clockOrigin =
        patch.clockDriftPerDayMs !== undefined &&
        patch.clockDriftPerDayMs !== state.faults.clockDriftPerDayMs
          ? Date.now()
          : state.clockOrigin;
      return { faults, clockOrigin };
    }),

  setIdentity: (patch) => set((state) => ({ identity: { ...state.identity, ...patch } })),

  patchGateway: (patch) =>
    set((state) => {
      const gateway = { ...state.gateway, ...patch };
      // Any schedule change re-plans the next poll from now, so an operator
      // switching to aligned mode sees the phase snap immediately instead of
      // waiting out the old interval.
      const rescheduled =
        patch.scheduleMode !== undefined ||
        patch.intervalMs !== undefined ||
        patch.offsetMs !== undefined;
      return {
        gateway,
        nextPollAt: rescheduled && state.running
          ? nextPollAt(Date.now(), gateway.scheduleMode, gateway.intervalMs, gateway.offsetMs)
          : state.nextPollAt,
      };
    }),

  start: () =>
    set((state) => {
      if (state.running) return state;
      const now = Date.now();
      const { scheduleMode, intervalMs, offsetMs } = state.gateway;
      return {
        running: true,
        connected: true,
        counters: { ...state.counters, connections: state.counters.connections + 1 },
        // Interval mode polls on connect; aligned mode waits for the boundary,
        // because polling immediately is exactly the phase error alignment is
        // there to remove.
        nextPollAt: scheduleMode === "interval" ? now : nextPollAt(now, scheduleMode, intervalMs, offsetMs),
        events: appendEvent(state.events, {
          id: nextEventId(),
          at: now,
          direction: "sys",
          source: "gateway",
          command: "connect",
          bytes: [],
          note: "pollingStarted",
          cycle: state.cycle,
          attempt: 0,
        }),
      };
    }),

  stop: () =>
    set((state) => {
      if (!state.running) return state;
      return {
        running: false,
        connected: false,
        identityRead: null,
        events: appendEvent(state.events, {
          id: nextEventId(),
          at: Date.now(),
          direction: "sys",
          source: "gateway",
          command: "disconnect",
          bytes: [],
          note: "pollingStopped",
          cycle: state.cycle,
          attempt: 0,
        }),
      };
    }),

  clearEvents: () => set({ events: [] }),

  reset: () =>
    set({
      running: false,
      connected: false,
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
      clockOrigin: Date.now(),
    }),

  tick: (now) => {
    const state = get();
    if (!state.running || now < state.nextPollAt) return;
    set(runCycle(state, now));
  },
}));

function appendEvent(events: BenchEvent[], event: BenchEvent): BenchEvent[] {
  const next = [...events, event];
  return next.length > MAX_EVENTS ? next.slice(next.length - MAX_EVENTS) : next;
}

function appendEvents(events: BenchEvent[], added: BenchEvent[]): BenchEvent[] {
  const next = [...events, ...added];
  return next.length > MAX_EVENTS ? next.slice(next.length - MAX_EVENTS) : next;
}

/**
 * One polling cycle, including in-session retries.
 *
 * The retry rule is the interesting part and mirrors the gateway exactly: a
 * frame that arrived corrupt, or did not arrive in time, is retried on the
 * *same* connection, while a connection the device dropped forces a reconnect.
 * Conflating the two is what turns a noisy line into a reconnect storm.
 */
function runCycle(state: BenchState, now: number): Partial<BenchState> {
  const { faults, gateway, checksumMode, adapterAddress } = state;
  const cycle = state.cycle + 1;
  const added: BenchEvent[] = [];
  const counters = { ...state.counters };

  let connected = state.connected;
  let health = state.health;
  let samples = state.samples;
  let lastDeviceTime = state.lastDeviceTime;
  let identityRead = state.identityRead;
  let cursor = now;

  // The identity probe runs once per connection and never fails a poll: an
  // older device that answers with a generic ACK is remembered as not
  // supporting the command rather than being retried forever.
  if (gateway.identityProbe && identityRead === null && !faults.noResponse) {
    const request = encodeFrame(0x00, adapterAddress, buildReadIdentityRequest(), checksumMode);
    added.push({
      id: nextEventId(),
      at: cursor,
      direction: "tx",
      source: "gateway",
      command: "read-identity",
      bytes: request,
      cycle,
      attempt: 0,
    });
    const latency = BASE_LATENCY_MS + Math.random() * JITTER_MS + faults.responseDelayMs;
    cursor += latency;
    const response = encodeFrame(
      0x00 | RESPONSE_BIT,
      adapterAddress,
      buildReadIdentityResponse(state.identity.model, state.identity.serial, state.identity.firmware),
      checksumMode,
    );
    added.push({
      id: nextEventId(),
      at: cursor,
      direction: "rx",
      source: "device",
      command: "read-identity",
      bytes: response,
      cycle,
      attempt: 0,
      latencyMs: latency,
    });
    identityRead = { ...state.identity };
  }

  let succeeded = false;
  let lastLatency = 0;

  for (let attempt = 0; attempt <= gateway.retryAttempts; attempt += 1) {
    if (attempt > 0) {
      counters.retries += 1;
      cursor += 200 * 2 ** (attempt - 1);
    }

    const request = encodeFrame(0x00, adapterAddress, buildReadTimeRequest(), checksumMode);
    const requestedAt = cursor;
    added.push({
      id: nextEventId(),
      at: requestedAt,
      direction: "tx",
      source: "gateway",
      command: "read-time",
      bytes: request,
      cycle,
      attempt,
    });

    if (faults.closeAfterRequest) {
      cursor += BASE_LATENCY_MS;
      added.push({
        id: nextEventId(),
        at: cursor,
        direction: "err",
        source: "device",
        command: "read-time",
        bytes: [],
        errorCode: "connectionClosed",
        cycle,
        attempt,
      });
      counters.reconnects += 1;
      connected = false;
      identityRead = null;
      break;
    }

    if (faults.noResponse) {
      cursor += gateway.requestTimeoutMs;
      added.push({
        id: nextEventId(),
        at: cursor,
        direction: "err",
        source: "gateway",
        command: "read-time",
        bytes: [],
        errorCode: "noResponse",
        cycle,
        attempt,
      });
      continue;
    }

    const fragmented = Math.random() < faults.fragmentProb;
    const latency =
      BASE_LATENCY_MS +
      Math.random() * JITTER_MS +
      faults.responseDelayMs +
      (fragmented ? faults.fragmentDelayMs : 0);

    if (latency > gateway.requestTimeoutMs) {
      cursor += gateway.requestTimeoutMs;
      added.push({
        id: nextEventId(),
        at: cursor,
        direction: "err",
        source: "gateway",
        command: "read-time",
        bytes: [],
        errorCode: "timeout",
        cycle,
        attempt,
      });
      continue;
    }

    cursor += latency;
    lastLatency = latency;

    const deviceTime = deviceClock(cursor, faults, state.clockOrigin);
    const response = encodeFrame(
      0x00 | RESPONSE_BIT,
      adapterAddress,
      buildReadTimeResponse(new Date(deviceTime)),
      checksumMode,
    );

    if (Math.random() < faults.badChecksumProb) {
      const corrupt = [...response];
      corrupt[corrupt.length - 2] = (corrupt[corrupt.length - 2]! ^ 0xff) & 0xff;
      added.push({
        id: nextEventId(),
        at: cursor,
        direction: "rx",
        source: "device",
        command: "read-time",
        bytes: corrupt,
        cycle,
        attempt,
        latencyMs: latency,
      });
      added.push({
        id: nextEventId(),
        at: cursor,
        direction: "err",
        source: "gateway",
        command: "read-time",
        bytes: [],
        errorCode: "invalidChecksum",
        cycle,
        attempt,
      });
      counters.protocolErrors += 1;
      continue;
    }

    added.push({
      id: nextEventId(),
      at: cursor,
      direction: "rx",
      source: "device",
      command: "read-time",
      bytes: response,
      cycle,
      attempt,
      latencyMs: latency,
    });

    // The wire carries whole seconds only, so the gateway sees a truncated
    // timestamp — that quantisation is part of the skew it measures and the
    // bench would flatter itself by hiding it.
    const wireTime = Math.floor(deviceTime / 1000) * 1000;
    const sample = sampleSkew(requestedAt, cursor, wireTime);
    samples = [...samples, sample].slice(-MAX_SAMPLES);
    lastDeviceTime = wireTime;
    succeeded = true;
    break;
  }

  if (succeeded) {
    counters.successfulReads += 1;
    health = recordOutcome(health, true, gateway.policy, lastLatency);
  } else {
    counters.failedReads += 1;
    if (!faults.closeAfterRequest) counters.exhaustedPolls += 1;
    health = recordOutcome(health, false, gateway.policy);
  }

  const previousState = state.health.state;
  if (health.state !== previousState) {
    added.push({
      id: nextEventId(),
      at: cursor,
      direction: "sys",
      source: "gateway",
      command: "device-state",
      bytes: [],
      note: "deviceStateChanged",
      noteArgs: { from: previousState, to: health.state },
      cycle,
      attempt: 0,
    });
  }

  const previousClock = buildClockReport(state.samples, gateway.thresholds).state;
  const currentClock = buildClockReport(samples, gateway.thresholds).state;
  if (currentClock !== previousClock) {
    added.push({
      id: nextEventId(),
      at: cursor,
      direction: "sys",
      source: "gateway",
      command: "clock-state",
      bytes: [],
      note: "clockStateChanged",
      noteArgs: { from: previousClock, to: currentClock },
      cycle,
      attempt: 0,
    });
  }

  // A dropped connection is re-established on the next tick after a short
  // backoff rather than instantly, so the reconnect is visible on the timeline.
  const reconnecting = !connected;
  if (reconnecting) {
    added.push({
      id: nextEventId(),
      at: cursor + 300,
      direction: "sys",
      source: "gateway",
      command: "reconnect",
      bytes: [],
      note: "reconnected",
      cycle,
      attempt: 0,
    });
    connected = true;
    counters.connections += 1;
  }

  return {
    connected,
    cycle,
    counters,
    health,
    samples,
    lastDeviceTime,
    identityRead,
    lastCycleAt: cursor,
    events: appendEvents(state.events, added),
    nextPollAt: nextPollAt(
      Math.max(now, cursor),
      gateway.scheduleMode,
      gateway.intervalMs,
      gateway.offsetMs,
    ),
  };
}

/* ------------------------------------------------------------- selectors */

/**
 * Derived clock report.
 *
 * A hook rather than a plain selector: `buildClockReport` allocates, and a
 * selector that returns a fresh object every call makes zustand's snapshot
 * comparison fail on every render — React reports that as an infinite loop.
 * Selecting the two stable inputs and memoising keeps the identity stable
 * until a sample actually arrives.
 */
export function useClockReport(): ClockReport {
  const samples = useBenchStore((state) => state.samples);
  const thresholds = useBenchStore((state) => state.gateway.thresholds);
  return useMemo(() => buildClockReport(samples, thresholds), [samples, thresholds]);
}

export function selectAvailability(state: BenchState): number {
  return availability(state.health);
}

export function selectLastFrameHex(state: BenchState): string {
  const last = [...state.events].reverse().find((event) => event.bytes.length > 0);
  return last ? formatHex(last.bytes) : "";
}

export function selectActiveFaultCount(state: BenchState): number {
  const { faults } = state;
  return [
    faults.responseDelayMs > 0,
    faults.badChecksumProb > 0,
    faults.fragmentProb > 0,
    faults.noResponse,
    faults.closeAfterRequest,
    faults.clockOffsetMs !== 0,
    faults.clockDriftPerDayMs !== 0,
  ].filter(Boolean).length;
}
