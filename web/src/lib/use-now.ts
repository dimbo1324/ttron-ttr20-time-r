"use client";

import { useEffect, useState } from "react";

/**
 * The current time, ticking.
 *
 * Countdowns and "time since" readouts cannot be driven by the store: it only
 * changes when a poll resolves, so a countdown between polls would sit still
 * for the whole interval. Both the dashboard and the gateway panel grew the
 * same effect; this is that effect, once.
 *
 * It starts at `0` rather than at `Date.now()` so the server render and the
 * first client render agree — the real value only arrives after mount, and
 * callers render a placeholder until then. `0` is falsy, which makes the guard
 * at the call site read naturally.
 */
export function useNow(intervalMs = 250): number {
  const [now, setNow] = useState(0);

  useEffect(() => {
    setNow(Date.now());
    const id = window.setInterval(() => setNow(Date.now()), intervalMs);
    return () => window.clearInterval(id);
  }, [intervalMs]);

  return now;
}
