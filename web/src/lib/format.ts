/**
 * Number and duration formatting for an instrument readout.
 *
 * The rule everywhere below: a value the operator compares against a threshold
 * keeps its unit and its sign. A skew of "-1.4 s" and one of "1.4 s" mean
 * opposite faults (device behind vs ahead), so the sign is never dropped.
 */

/** Compact duration: 940 ms, 1.42 s, 2 м 05 с. */
export function formatDuration(ms: number, options: { signed?: boolean } = {}): string {
  if (!Number.isFinite(ms)) return "—";
  // Exactly zero is a reading, not a measurement to two decimal places.
  if (ms === 0) return "0 ms";

  const sign = ms < 0 ? "-" : options.signed && ms > 0 ? "+" : "";
  const value = Math.abs(ms);

  if (value < 1) return `${sign}${value.toFixed(2)} ms`;
  if (value < 1000) return `${sign}${Math.round(value)} ms`;
  if (value < 60_000) return `${sign}${(value / 1000).toFixed(value < 10_000 ? 2 : 1)} s`;

  const totalSeconds = Math.round(value / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  if (minutes < 60) return `${sign}${minutes}m ${String(seconds).padStart(2, "0")}s`;

  const hours = Math.floor(minutes / 60);
  return `${sign}${hours}h ${String(minutes % 60).padStart(2, "0")}m`;
}

/** Drift is quoted per day and is small enough to want two decimals. */
export function formatDriftPerDay(msPerDay: number): string {
  if (!Number.isFinite(msPerDay) || msPerDay === 0) return "0 s";
  const seconds = msPerDay / 1000;
  const sign = seconds > 0 ? "+" : "-";
  const value = Math.abs(seconds);
  if (value < 1) return `${sign}${(value * 1000).toFixed(0)} ms`;
  if (value < 60) return `${sign}${value.toFixed(value < 10 ? 2 : 1)} s`;
  return `${sign}${(value / 60).toFixed(1)} m`;
}

export function formatPercent(ratio: number, digits = 1): string {
  if (!Number.isFinite(ratio)) return "—";
  return `${(ratio * 100).toFixed(digits)}%`;
}

/** Wall clock with milliseconds — the log needs sub-second ordering. */
export function formatClock(at: number, withMillis = true): string {
  const date = new Date(at);
  const pad = (value: number, width = 2) => String(value).padStart(width, "0");
  const base = `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
  return withMillis ? `${base}.${pad(date.getMilliseconds(), 3)}` : base;
}

export function formatDate(at: number): string {
  const date = new Date(at);
  const pad = (value: number) => String(value).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
}

export function formatCount(value: number): string {
  return new Intl.NumberFormat("ru-RU").format(value);
}
