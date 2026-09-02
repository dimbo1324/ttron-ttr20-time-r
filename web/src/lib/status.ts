import type { ClockState, HealthState } from "@/lib/bench/domain";

/**
 * How a domain state is coloured, in one place.
 *
 * The mapping lived inline in four components — the shell strip, both
 * dashboard cards and the gateway panel — as a chain of ternaries, and the
 * chains had already started to disagree about what `unknown` looks like.
 *
 * `Record<ClockState, …>` is what makes this worth extracting rather than
 * merely tidy: adding a state to the domain now fails to compile until every
 * map here has an entry for it, so a new state cannot reach the screen
 * uncoloured and unnamed.
 */

/** Tones the `Badge` primitive accepts for a state. */
export type StateTone = "neutral" | "success" | "warning" | "danger";

/** Tones the `StatusDot` primitive accepts. */
export type DotTone = "online" | "degraded" | "offline" | "idle" | "unknown";

/**
 * Every state the two machines can be in.
 *
 * Both unions are keys of the dictionary's `states` section, which is what
 * lets a label be looked up without a cast.
 */
export type DomainState = ClockState | HealthState;

export const CLOCK_TONE = {
  unknown: "neutral",
  ok: "success",
  warn: "warning",
  critical: "danger",
} as const satisfies Record<ClockState, StateTone>;

export const CLOCK_DOT = {
  unknown: "unknown",
  ok: "online",
  warn: "degraded",
  critical: "offline",
} as const satisfies Record<ClockState, DotTone>;

export const HEALTH_TONE = {
  unknown: "neutral",
  online: "success",
  degraded: "warning",
  offline: "danger",
} as const satisfies Record<HealthState, StateTone>;

export const HEALTH_DOT = {
  unknown: "unknown",
  online: "online",
  degraded: "degraded",
  offline: "offline",
} as const satisfies Record<HealthState, DotTone>;

/** Direction of a frame on the wire, shared by the log and the diagram. */
export const DIRECTION_TONE = {
  tx: "tx",
  rx: "rx",
  err: "err",
  sys: "sys",
} as const;

export const DIRECTION_TEXT = {
  tx: "text-signal-tx",
  rx: "text-signal-rx",
  err: "text-signal-err",
  sys: "text-signal-sys",
} as const;

export const DIRECTION_BG = {
  tx: "bg-signal-tx",
  rx: "bg-signal-rx",
  err: "bg-signal-err",
  sys: "bg-signal-sys",
} as const;
