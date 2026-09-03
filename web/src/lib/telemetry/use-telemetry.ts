"use client";

import { useEffect, useMemo } from "react";

import { useBenchStore, useClockReport, type DeviceFaults } from "@/stores/bench-store";
import { liveTelemetry, useLiveStore } from "@/stores/live-store";
import { useSourceStore } from "@/stores/source-store";
import type { ApiFaultMode } from "@/lib/api/schema";

import { benchTelemetry } from "./from-bench";
import type { FaultView, Telemetry, TelemetrySource } from "./types";

/**
 * The one place a component asks what is happening.
 *
 * Both sources are subscribed to unconditionally — hooks cannot be called
 * behind a branch — and the inactive one is simply not returned. That costs a
 * subscription to a store that is not changing, which is nothing, and buys the
 * rule that a panel never learns which source it is rendering.
 */
export function useTelemetry(): Telemetry {
  const source = useSourceStore((state) => state.source);

  const bench = useBenchStore();
  const clock = useClockReport();
  const live = useLiveStore();

  const benchView = useMemo(() => benchTelemetry(bench, clock), [bench, clock]);
  const liveView = useMemo(() => liveTelemetry(live), [live]);

  return source === "live" ? liveView : benchView;
}

export function useSource(): TelemetrySource {
  return useSourceStore((state) => state.source);
}

/** How often the live source re-reads the gateway, in ms. */
const LIVE_INTERVAL_MS = 1000;

/**
 * How often the bench engine is advanced, in ms.
 *
 * Well below the fastest schedule the UI offers (1s) and well above the rate
 * at which re-rendering a live log becomes the bottleneck. The store itself
 * no-ops until the next poll instant, so a tick that lands early costs one
 * comparison.
 */
const BENCH_TICK_MS = 120;

/**
 * Drives whichever source is selected. Mounted once, by the shell.
 *
 * Keeping both loops here rather than inside their stores is what guarantees
 * only one of them is ever running: switching source tears down one effect and
 * sets up the other, and a bench that keeps ticking behind a live view is
 * exactly the kind of thing that shows up later as a mysterious event in the
 * log.
 */
export function useTelemetryEngine(): void {
  const source = useSourceStore((state) => state.source);
  const hydrate = useSourceStore((state) => state.hydrate);

  const running = useBenchStore((state) => state.running);
  const tick = useBenchStore((state) => state.tick);

  const refresh = useLiveStore((state) => state.refresh);
  const loadFaults = useLiveStore((state) => state.loadFaults);

  // The stored source can only be read in the browser; see source-store.
  useEffect(() => hydrate(), [hydrate]);

  useEffect(() => {
    if (source !== "bench" || !running) return;
    const id = window.setInterval(() => tick(Date.now()), BENCH_TICK_MS);
    return () => window.clearInterval(id);
  }, [source, running, tick]);

  useEffect(() => {
    if (source !== "live") return;

    void refresh();
    void loadFaults();
    const id = window.setInterval(() => void refresh(), LIVE_INTERVAL_MS);
    return () => window.clearInterval(id);
  }, [source, refresh, loadFaults]);
}

/**
 * Start and stop, routed to whichever source is selected.
 *
 * `reset` is bench-only on purpose: clearing counters on the live source would
 * mean asking a running gateway to forget what it has measured, which is not
 * something a console should be able to do from a toolbar.
 */
export function useTelemetryControls(): {
  start: () => void;
  stop: () => void;
  reset: (() => void) | null;
  busy: boolean;
} {
  const source = useSourceStore((state) => state.source);

  const benchStart = useBenchStore((state) => state.start);
  const benchStop = useBenchStore((state) => state.stop);
  const benchReset = useBenchStore((state) => state.reset);

  const liveStart = useLiveStore((state) => state.start);
  const liveStop = useLiveStore((state) => state.stop);
  const busy = useLiveStore((state) => state.busy);

  if (source === "live") {
    return { start: () => void liveStart(), stop: () => void liveStop(), reset: null, busy };
  }
  return { start: benchStart, stop: benchStop, reset: benchReset, busy: false };
}

/**
 * Fault injection, routed to whichever source is selected.
 *
 * The patch is written in the console's own vocabulary and translated at the
 * edge, so a panel never has to know that the Go emulator spells a corrupt
 * checksum as a flag plus a probability while the bench spells it as one
 * number. `writable` is false when the live emulator has not answered — an
 * operator dragging a slider that cannot reach anything deserves to be told,
 * not to watch it spring back.
 */
export function useFaultControls(): {
  setFaults: (patch: Partial<FaultView>) => void;
  writable: boolean;
} {
  const source = useSourceStore((state) => state.source);

  const patchBench = useBenchStore((state) => state.patchFaults);
  const setLive = useLiveStore((state) => state.setFaults);
  const liveFaults = useLiveStore((state) => state.faults);

  if (source === "live") {
    return {
      writable: liveFaults !== null,
      setFaults: (patch) => void setLive(toApiFaults(patch)),
    };
  }
  return {
    writable: true,
    // The bench cannot be sent a null clock offset -- that combination only
    // exists to say "the live source has no such control" -- so the nulls are
    // dropped rather than written through.
    setFaults: (patch) => patchBench(toBenchFaults(patch)),
  };
}

/**
 * The console's fault vocabulary, in the emulator's.
 *
 * A probability of exactly 1 sets the emulator's boolean as well, because that
 * is what "every frame" means to it; the clock fields have no counterpart and
 * are simply dropped.
 */
function toBenchFaults(patch: Partial<FaultView>): Partial<DeviceFaults> {
  const { clockOffsetMs, clockDriftPerDayMs, ...rest } = patch;
  return {
    ...rest,
    ...(clockOffsetMs !== null && clockOffsetMs !== undefined ? { clockOffsetMs } : {}),
    ...(clockDriftPerDayMs !== null && clockDriftPerDayMs !== undefined
      ? { clockDriftPerDayMs }
      : {}),
  };
}

function toApiFaults(patch: Partial<FaultView>): Partial<ApiFaultMode> {
  const out: Partial<ApiFaultMode> = {};
  if (patch.responseDelayMs !== undefined) out.responseDelayMs = patch.responseDelayMs;
  if (patch.fragmentDelayMs !== undefined) out.fragmentDelayMs = patch.fragmentDelayMs;
  if (patch.noResponse !== undefined) out.noResponse = patch.noResponse;
  if (patch.closeAfterRequest !== undefined) out.closeAfterRequest = patch.closeAfterRequest;
  if (patch.badChecksumProb !== undefined) {
    out.corruptChecksumProbability = patch.badChecksumProb;
    out.corruptChecksum = patch.badChecksumProb >= 1;
  }
  if (patch.fragmentProb !== undefined) {
    out.fragmentProbability = patch.fragmentProb;
    out.fragmentResponse = patch.fragmentProb >= 1;
  }
  return out;
}
