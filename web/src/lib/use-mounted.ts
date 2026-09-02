"use client";

import { useEffect, useState } from "react";

/**
 * True once the component has mounted in the browser.
 *
 * Anything derived from the current time cannot be rendered on the server and
 * on the client and be expected to match — the two run at different instants,
 * and React reports the difference as a hydration failure. Readouts that show
 * "now" therefore render a placeholder until this flips.
 */
export function useMounted(): boolean {
  const [mounted, setMounted] = useState(false);
  useEffect(() => setMounted(true), []);
  return mounted;
}
