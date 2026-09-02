import { render, type RenderOptions, type RenderResult } from "@testing-library/react";
import type { ReactElement, ReactNode } from "react";

import { LocaleProvider } from "@/components/locale-provider";
import { getDictionary, type Locale } from "@/i18n";
import { EMPTY_HEALTH } from "@/lib/bench/domain";
import { DEFAULT_FAULTS, DEFAULT_GATEWAY, useBenchStore } from "@/stores/bench-store";

type BenchStoreState = ReturnType<typeof useBenchStore.getState>;

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
