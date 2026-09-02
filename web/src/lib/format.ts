/**
 * Readouts, in the reader's language.
 *
 * Two things here are locale-dependent and were not before: the unit symbol
 * (a Russian engineering readout says "мс", not "ms") and the decimal
 * separator (Russian writes 1,42 where English writes 1.42). Both are visible
 * on every panel, and getting them wrong is the difference between a console
 * that was translated and one that was localised.
 *
 * Everything is bound once through `createFormatter` rather than threaded as
 * a parameter through twenty call sites, and the components reach it through
 * `useFormat`.
 *
 * The rule that survives translation: a value the operator compares against a
 * threshold keeps its unit and its sign. A skew of "-1,4 с" and one of "1,4 с"
 * mean opposite faults — the device behind the reference or ahead of it — so
 * the sign is never dropped.
 */

export interface FormatUnits {
  ms: string;
  s: string;
  m: string;
  h: string;
  perDay: string;
}

export interface Formatter {
  /** Compact duration: 940 мс, 1,42 с, 2 м 05 с. */
  duration(ms: number, options?: { signed?: boolean }): string;
  /** Drift quoted per day, always signed — its direction is the whole point. */
  driftPerDay(msPerDay: number): string;
  percent(ratio: number, digits?: number): string;
  count(value: number): string;
  /** Wall clock with milliseconds — the log needs sub-second ordering. */
  clock(at: number, withMillis?: boolean): string;
  date(at: number): string;
}

/** Fallback for callers with no dictionary to hand: SI symbols, English. */
export const SI_UNITS: FormatUnits = { ms: "ms", s: "s", m: "m", h: "h", perDay: "/day" };

const pad = (value: number, width = 2) => String(value).padStart(width, "0");

export function createFormatter(locale: string, units: FormatUnits = SI_UNITS): Formatter {
  const number = (value: number, digits: number) =>
    new Intl.NumberFormat(locale, {
      minimumFractionDigits: digits,
      maximumFractionDigits: digits,
    }).format(value);

  const duration: Formatter["duration"] = (ms, options = {}) => {
    if (!Number.isFinite(ms)) return "—";
    // Exactly zero is a reading, not a measurement to two decimal places.
    if (ms === 0) return `0 ${units.ms}`;

    const sign = ms < 0 ? "-" : options.signed ? "+" : "";
    const value = Math.abs(ms);

    if (value < 1) return `${sign}${number(value, 2)} ${units.ms}`;
    if (value < 1000) return `${sign}${number(Math.round(value), 0)} ${units.ms}`;
    if (value < 60_000) {
      return `${sign}${number(value / 1000, value < 10_000 ? 2 : 1)} ${units.s}`;
    }

    const totalSeconds = Math.round(value / 1000);
    const minutes = Math.floor(totalSeconds / 60);
    if (minutes < 60) {
      return `${sign}${minutes} ${units.m} ${pad(totalSeconds % 60)} ${units.s}`;
    }
    return `${sign}${Math.floor(minutes / 60)} ${units.h} ${pad(minutes % 60)} ${units.m}`;
  };

  return {
    duration,

    driftPerDay(msPerDay) {
      if (!Number.isFinite(msPerDay) || msPerDay === 0) return `0 ${units.s}${units.perDay}`;

      const seconds = msPerDay / 1000;
      const sign = seconds > 0 ? "+" : "-";
      const value = Math.abs(seconds);

      const magnitude =
        value < 1
          ? `${number(Math.round(value * 1000), 0)} ${units.ms}`
          : value < 60
            ? `${number(value, value < 10 ? 2 : 1)} ${units.s}`
            : `${number(value / 60, 1)} ${units.m}`;

      return `${sign}${magnitude}${units.perDay}`;
    },

    percent(ratio, digits = 1) {
      if (!Number.isFinite(ratio)) return "—";
      return new Intl.NumberFormat(locale, {
        style: "percent",
        minimumFractionDigits: digits,
        maximumFractionDigits: digits,
      }).format(ratio);
    },

    count(value) {
      return new Intl.NumberFormat(locale).format(value);
    },

    clock(at, withMillis = true) {
      const date = new Date(at);
      const base = `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
      return withMillis ? `${base}.${pad(date.getMilliseconds(), 3)}` : base;
    },

    date(at) {
      const date = new Date(at);
      // Deliberately ISO in both locales: this is a machine timestamp, and the
      // one thing a reader must never have to guess is whether 03-09 is
      // March or September.
      return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
    },
  };
}
