"use client";

import { create } from "zustand";

import type { TelemetrySource } from "@/lib/telemetry/types";

/**
 * Which stack the console is looking at.
 *
 * Kept in its own store rather than folded into either engine, because both
 * engines have to be able to ask the question without depending on each other:
 * the live poller stops when the bench is selected, and the bench ticker stops
 * when the live source is.
 *
 * The choice is remembered per browser. Someone who has a Go stack running
 * wants the console to come back up pointed at it, and someone teaching with
 * the bench wants the same in reverse; making them re-pick on every reload is
 * the kind of small friction that ends with the wrong source being read as
 * real.
 */

const STORAGE_KEY = "ft12.source";

/**
 * Storage is read defensively: it throws outright in a browser configured to
 * block site data, and the console has to come up anyway.
 */
export function readStoredSource(): TelemetrySource {
  try {
    return window.localStorage.getItem(STORAGE_KEY) === "live" ? "live" : "bench";
  } catch {
    return "bench";
  }
}

function storeSource(source: TelemetrySource): void {
  try {
    window.localStorage.setItem(STORAGE_KEY, source);
  } catch {
    // A preference that cannot be saved is not worth failing a click over.
  }
}

interface SourceState {
  source: TelemetrySource;
  /**
   * True once the stored preference has been applied.
   *
   * The server cannot know what this browser chose, so the first client render
   * has to match the server's — `bench` — and only then adopt the stored
   * value. Rendering the stored source immediately would be a hydration
   * mismatch on every reload for anyone who picked `live`.
   */
  hydrated: boolean;
  setSource: (source: TelemetrySource) => void;
  hydrate: () => void;
}

export const useSourceStore = create<SourceState>()((set, get) => ({
  source: "bench",
  hydrated: false,

  setSource: (source) => {
    if (get().source === source) return;
    storeSource(source);
    set({ source });
  },

  hydrate: () => {
    if (get().hydrated) return;
    set({ source: readStoredSource(), hydrated: true });
  },
}));
