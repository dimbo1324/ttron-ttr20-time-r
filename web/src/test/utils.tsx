import { render, type RenderOptions, type RenderResult } from "@testing-library/react";
import type { ReactElement, ReactNode } from "react";

import { LocaleProvider } from "@/components/locale-provider";
import { getDictionary, type Locale } from "@/i18n";
import { EMPTY_HEALTH } from "@/lib/bench/domain";
import type { TelemetrySource } from "@/lib/telemetry/types";
import { DEFAULT_FAULTS, DEFAULT_GATEWAY, useBenchStore } from "@/stores/bench-store";
import { useLiveStore } from "@/stores/live-store";
import { useSourceStore } from "@/stores/source-store";

type BenchStoreState = ReturnType<typeof useBenchStore.getState>;
type LiveStoreState = ReturnType<typeof useLiveStore.getState>;

/**
 * Renders inside the dictionary provider every component in this app expects.
 *
 * Tests assert against the Russian copy because Russian is the primary locale
 * and the one the shape of every other dictionary is checked against — a
 * string that changes there is a change worth a failing test.
 */
export function renderWithLocale(
  ui: ReactElement,
  { locale = "ru" as Locale, ...options }: RenderOptions & { locale?: Locale } = {},
): RenderResult & { dict: ReturnType<typeof getDictionary> } {
  const dict = getDictionary(locale);

  const Wrapper = ({ children }: { children: ReactNode }) => (
    <LocaleProvider dict={dict} locale={locale}>
      {children}
    </LocaleProvider>
  );

  return { ...render(ui, { wrapper: Wrapper, ...options }), dict };
}

export { getDictionary };

/**
 * Returns the bench engine to its initial state.
 *
 * The store is a module singleton, so without this a test inherits whatever
 * the previous one polled — and the screens under test read it directly.
 */
export function resetBenchStore(overrides: Partial<BenchStoreState> = {}) {
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
    clockOrigin: Date.now(),
    ...overrides,
  });
}

/**
 * Intl separates a number from its unit with a non-breaking space, and
 * Testing Library normalises whitespace in the DOM before matching. Expected
 * strings have to make the same trip or a correct assertion fails on an
 * invisible character.
 */
export function plain(value: string): string {
  return value.replace(/[\u00A0\u202F]/g, " ");
}

/**
 * Returns the live source to its initial state.
 *
 * Same reasoning as `resetBenchStore`: both stores are module singletons, and
 * a test that leaves a status behind makes the next one pass for the wrong
 * reason. `reset()` is the store's own action, so this cannot drift from it.
 */
export function resetLiveStore(overrides: Partial<LiveStoreState> = {}) {
  useLiveStore.getState().reset();
  if (Object.keys(overrides).length > 0) useLiveStore.setState(overrides);
}

/** Selects a telemetry source for the components under test. */
export function useSource(source: TelemetrySource) {
  useSourceStore.setState({ source, hydrated: true });
}

/**
 * A gateway status with every section filled in.
 *
 * Written as the JSON the API actually returns rather than as a parsed object,
 * so a schema that stops accepting a real payload fails here.
 */
export function gatewayStatusFixture(overrides: Record<string, unknown> = {}) {
  return {
    state: "running",
    targetAddr: "127.0.0.1:9000",
    checksumMode: "sum",
    pollingIntervalMs: 5000,
    requestTimeoutMs: 1500,
    connectTimeoutMs: 2000,
    connected: true,
    connectionAttempts: 1,
    successfulReads: 42,
    failedReads: 3,
    reconnects: 1,
    protocolErrors: 2,
    lastRoundTripMs: 8,
    lastSuccessfulReadTime: "2026-09-02T12:00:00Z",
    lastDeviceTime: "2026-09-02T11:59:58Z",
    lastError: "",
    recentFramesCount: 11,
    deviceId: "alpha",
    deviceName: "TTR20 alpha",
    schedule: {
      mode: "aligned",
      intervalMs: 60_000,
      offsetMs: 5000,
      description: "aligned every 1m0s at offset 5s",
      nextPollAt: "2026-09-02T12:01:05Z",
    },
    retry: { attempts: 2, delayMs: 200, maxDelayMs: 800, totalRetries: 7, exhaustedPolls: 1 },
    clock: {
      state: "warn",
      skewMs: -2500,
      medianSkewMs: -2400,
      minSkewMs: -3000,
      maxSkewMs: -1800,
      driftPerDayMs: 24_000,
      driftDetermined: true,
      driftFit: 0.94,
      samples: 31,
      warnThresholdMs: 2000,
      criticalThresholdMs: 30_000,
      roundTripMs: 9,
      updatedAt: "2026-09-02T12:00:00Z",
      observedSamples: 32,
      rejectedSamples: 1,
    },
    health: {
      state: "degraded",
      since: "2026-09-02T11:00:00Z",
      availability: 0.875,
      windowSamples: 40,
      consecutiveFailures: 4,
      consecutiveSuccesses: 0,
      latencyP50Ms: 44,
      latencyP95Ms: 45,
      latencyP99Ms: 46,
      latencyMaxMs: 47,
      latencyMeanMs: 48,
      degradeAfter: 3,
      offlineAfter: 10,
      recoverAfter: 2,
    },
    identity: {
      known: true,
      supported: true,
      model: "TTR20",
      serial: "SN-42",
      firmware: "1.2.3",
      readAt: "2026-09-02T11:59:00Z",
    },
    ...overrides,
  };
}

export function fleetFixture() {
  return {
    summary: {
      devices: 2,
      running: 2,
      online: 1,
      degraded: 1,
      offline: 0,
      unknown: 0,
      clockOk: 1,
      clockWarn: 1,
      clockCritical: 0,
      clockUnknown: 0,
      worstClockSkewMs: -2500,
      worstClockDeviceId: "alpha",
    },
    devices: [
      gatewayStatusFixture(),
      gatewayStatusFixture({
        deviceId: "beta",
        deviceName: "TTR20 beta",
        targetAddr: "127.0.0.1:9001",
        health: { ...gatewayStatusFixture().health, state: "online", availability: 1 },
        clock: { ...gatewayStatusFixture().clock, state: "ok", skewMs: 12, samples: 9 },
      }),
    ],
  };
}

export function historyFixture() {
  return {
    clockSamples: [
      { at: "2026-09-02T11:59:00Z", skewMs: -2600, roundTripMs: 8 },
      { at: "2026-09-02T12:00:00Z", skewMs: -2500, roundTripMs: 9 },
    ],
    healthOutcomes: [
      { at: "2026-09-02T11:59:00Z", success: true, latencyMs: 8 },
      { at: "2026-09-02T12:00:00Z", success: false, latencyMs: 0 },
    ],
  };
}

export function faultModeFixture(overrides: Record<string, unknown> = {}) {
  return {
    responseDelayMs: 0,
    corruptChecksum: false,
    corruptChecksumProbability: 0,
    fragmentResponse: false,
    fragmentProbability: 0,
    fragmentDelayMs: 40,
    noResponse: false,
    closeAfterRequest: false,
    ...overrides,
  };
}

/** A gateway frame record, as `/gateway/events` returns it. */
export function eventFixture(overrides: Record<string, unknown> = {}) {
  return {
    id: 1,
    timestamp: "2026-09-02T12:00:00.000Z",
    source: "gateway",
    service: "gateway",
    direction: "TX",
    remoteAddr: "127.0.0.1:9000",
    checksumMode: "sum",
    rawHex: "68 03 68 00 01 01 02 16",
    command: "read-time",
    error: "",
    message: "",
    ...overrides,
  };
}

/** The settings the API returns and accepts, matching gatewayStatusFixture. */
export function settingsFixture(overrides: Record<string, unknown> = {}) {
  return {
    scheduleMode: "aligned",
    pollIntervalMs: 60_000,
    pollOffsetMs: 5000,
    requestTimeoutMs: 1500,
    retryAttempts: 2,
    retryDelayMs: 200,
    clockWarnMs: 2000,
    clockCriticalMs: 30_000,
    degradeAfter: 3,
    offlineAfter: 10,
    recoverAfter: 2,
    ...overrides,
  };
}
