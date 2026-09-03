import { z } from "zod";

/**
 * The shapes the Go HTTP API returns, as parsers rather than as types.
 *
 * Everywhere else in this console a `type` is enough, because the value was
 * produced by code in the same bundle. Here it is not: this data crosses a
 * process boundary, from a binary that can be older than the page talking to
 * it. A plain `as GatewayStatus` would turn a renamed field into `undefined`
 * and surface it four components later as `NaN ms` with nothing to point at.
 *
 * So every response is parsed. Numbers carry defaults so that a field the
 * running gateway does not know about yet degrades to a zero instead of
 * taking the whole page down, while a genuinely wrong shape — an object where
 * a list belongs — still fails loudly at the boundary, which is the one place
 * an operator can be told something useful about it.
 */

/** Milliseconds, defaulting to zero when absent. */
const ms = z.number().default(0);
const count = z.number().default(0);

/** RFC3339 instant, omitted by the API when the underlying time is unset. */
const instant = z.string().optional();

export const scheduleSchema = z.object({
  mode: z.string().default(""),
  intervalMs: ms,
  offsetMs: ms,
  description: z.string().default(""),
  nextPollAt: instant,
});

export const retrySchema = z.object({
  attempts: count,
  delayMs: ms,
  maxDelayMs: ms,
  totalRetries: count,
  exhaustedPolls: count,
});

export const clockSchema = z.object({
  state: z.string().default("unknown"),
  skewMs: ms,
  medianSkewMs: ms,
  minSkewMs: ms,
  maxSkewMs: ms,
  driftPerDayMs: ms,
  driftDetermined: z.boolean().default(false),
  driftFit: z.number().default(0),
  samples: count,
  warnThresholdMs: ms,
  criticalThresholdMs: ms,
  roundTripMs: ms,
  updatedAt: instant,
  observedSamples: count,
  rejectedSamples: count,
});

export const deviceHealthSchema = z.object({
  state: z.string().default("unknown"),
  since: instant,
  availability: z.number().default(0),
  windowSamples: count,
  consecutiveFailures: count,
  consecutiveSuccesses: count,
  latencyP50Ms: ms,
  latencyP95Ms: ms,
  latencyP99Ms: ms,
  latencyMaxMs: ms,
  latencyMeanMs: ms,
  degradeAfter: count,
  offlineAfter: count,
  recoverAfter: count,
});

export const identitySchema = z.object({
  known: z.boolean().default(false),
  supported: z.boolean().default(false),
  model: z.string().default(""),
  serial: z.string().default(""),
  firmware: z.string().default(""),
  readAt: instant,
});

export const gatewayStatusSchema = z.object({
  state: z.string().default("unspecified"),
  targetAddr: z.string().default(""),
  checksumMode: z.string().default("unspecified"),
  pollingIntervalMs: ms,
  requestTimeoutMs: ms,
  connectTimeoutMs: ms,
  connected: z.boolean().default(false),
  connectionAttempts: count,
  successfulReads: count,
  failedReads: count,
  reconnects: count,
  protocolErrors: count,
  lastRoundTripMs: ms,
  lastSuccessfulReadTime: instant,
  lastDeviceTime: instant,
  lastError: z.string().default(""),
  lastTxTime: instant,
  lastRxTime: instant,
  recentFramesCount: count,
  deviceId: z.string().default(""),
  deviceName: z.string().default(""),
  schedule: scheduleSchema.default({}),
  retry: retrySchema.default({}),
  clock: clockSchema.default({}),
  health: deviceHealthSchema.default({}),
  identity: identitySchema.default({}),
});

export const fleetSchema = z.object({
  summary: z
    .object({
      devices: count,
      running: count,
      online: count,
      degraded: count,
      offline: count,
      unknown: count,
      clockOk: count,
      clockWarn: count,
      clockCritical: count,
      clockUnknown: count,
      worstClockSkewMs: ms,
      worstClockDeviceId: z.string().default(""),
    })
    .default({}),
  devices: z.array(gatewayStatusSchema).default([]),
});

export const historySchema = z.object({
  clockSamples: z
    .array(z.object({ at: instant, skewMs: ms, roundTripMs: ms }))
    .default([]),
  healthOutcomes: z
    .array(z.object({ at: instant, success: z.boolean().default(false), latencyMs: ms }))
    .default([]),
});

export const eventSchema = z.object({
  id: z.number().default(0),
  timestamp: instant,
  source: z.string().default(""),
  service: z.string().default(""),
  direction: z.string().default("SYSTEM"),
  remoteAddr: z.string().default(""),
  checksumMode: z.string().default("unspecified"),
  rawHex: z.string().default(""),
  command: z.string().default(""),
  error: z.string().default(""),
  message: z.string().default(""),
});

export const eventsSchema = z.object({ events: z.array(eventSchema).default([]) });

export const faultModeSchema = z.object({
  responseDelayMs: ms,
  corruptChecksum: z.boolean().default(false),
  corruptChecksumProbability: z.number().default(0),
  fragmentResponse: z.boolean().default(false),
  fragmentProbability: z.number().default(0),
  fragmentDelayMs: ms,
  noResponse: z.boolean().default(false),
  closeAfterRequest: z.boolean().default(false),
});

export const emulatorStatusSchema = z.object({
  state: z.string().default("unspecified"),
  listenAddr: z.string().default(""),
  checksumMode: z.string().default("unspecified"),
  activeConnections: count,
  totalConnections: count,
  totalRequests: count,
  totalResponses: count,
  protocolErrors: count,
  lastError: z.string().default(""),
  recentFramesCount: count,
  faultMode: faultModeSchema.optional(),
});

/**
 * The part of a gateway's configuration that can change while it is running.
 *
 * Not parsed leniently on the way out: this is the one shape this console
 * *sends*, and a field defaulted to zero here would reconfigure a running
 * gateway to something nobody asked for. The gateway rejects it, which is the
 * behaviour these defaults exist to preserve on the way in.
 */
export const gatewaySettingsSchema = z.object({
  scheduleMode: z.string().default("interval"),
  pollIntervalMs: ms,
  pollOffsetMs: ms,
  requestTimeoutMs: ms,
  retryAttempts: count,
  retryDelayMs: ms,
  clockWarnMs: ms,
  clockCriticalMs: ms,
  degradeAfter: count,
  offlineAfter: count,
  recoverAfter: count,
});

/**
 * `PUT /gateway/settings` answers with what was applied and the status it
 * produced, so the console redraws without a second round trip.
 */
export const updateSettingsSchema = z.object({
  settings: gatewaySettingsSchema,
  status: gatewayStatusSchema,
});

/** `POST /gateway/start` and `/stop` both answer with the resulting status. */
export const gatewayCommandSchema = z.object({ status: gatewayStatusSchema });

/**
 * `PUT /emulator/fault-mode` answers with the accepted fault mode *and* the
 * emulator status. Only the first is read back here -- the status arrives on
 * the next poll anyway, and threading it through the setter would give the
 * caller two sources for one value.
 */
export const setFaultModeSchema = z.object({
  faultMode: faultModeSchema,
  status: emulatorStatusSchema.optional(),
});

export type ApiGatewayStatus = z.infer<typeof gatewayStatusSchema>;
export type ApiFleet = z.infer<typeof fleetSchema>;
export type ApiHistory = z.infer<typeof historySchema>;
export type ApiEvent = z.infer<typeof eventSchema>;
export type ApiFaultMode = z.infer<typeof faultModeSchema>;
export type ApiEmulatorStatus = z.infer<typeof emulatorStatusSchema>;
export type ApiGatewaySettings = z.infer<typeof gatewaySettingsSchema>;
