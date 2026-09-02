import { act, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { useBenchStore } from "@/stores/bench-store";
import { renderWithLocale, resetBenchStore } from "@/test/utils";

import { AppShell } from "./app-shell";

/** The shell reads the current route to mark the active section. */
let pathname = "/ru";
jest.mock("next/navigation", () => ({
  usePathname: () => pathname,
}));

beforeEach(() => {
  pathname = "/ru";
  resetBenchStore();
});

describe("AppShell navigation", () => {
  it("lists every section with a link that carries the locale", () => {
    const { dict } = renderWithLocale(
      <AppShell>
        <p>содержимое</p>
      </AppShell>,
    );

    const links: [string, string][] = [
      [dict.nav.overview, "/ru"],
      [dict.nav.monitor, "/ru/monitor"],
      [dict.nav.protocol, "/ru/protocol"],
      [dict.nav.emulator, "/ru/emulator"],
      [dict.nav.gateway, "/ru/gateway"],
      [dict.nav.reference, "/ru/reference"],
    ];

    for (const [label, href] of links) {
      expect(screen.getByRole("link", { name: label })).toHaveAttribute("href", href);
    }
  });

  it("groups the sections under their headings", () => {
    const { dict } = renderWithLocale(<AppShell>x</AppShell>);

    expect(screen.getByText(dict.nav.sectionWork)).toBeInTheDocument();
    expect(screen.getByText(dict.nav.sectionControl)).toBeInTheDocument();
    expect(screen.getByText(dict.nav.sectionLearn)).toBeInTheDocument();
  });

  it("renders its children", () => {
    renderWithLocale(
      <AppShell>
        <p>содержимое</p>
      </AppShell>,
    );

    expect(screen.getByText("содержимое")).toBeInTheDocument();
  });

  it("offers both locales, pointing at the current route", () => {
    pathname = "/ru/monitor";
    renderWithLocale(<AppShell>x</AppShell>);

    expect(screen.getByRole("link", { name: "EN" })).toHaveAttribute("href", "/en/monitor");
    expect(screen.getByRole("link", { name: "RU" })).toHaveAttribute("href", "/ru/monitor");
  });

  it("renders in English when that is the locale", () => {
    pathname = "/en";
    const { dict } = renderWithLocale(<AppShell>x</AppShell>, { locale: "en" });

    expect(screen.getByRole("link", { name: dict.nav.overview })).toHaveAttribute("href", "/en");
  });
});

describe("AppShell status strip", () => {
  it("shows the device state and an unmeasured clock", () => {
    const { dict } = renderWithLocale(<AppShell>x</AppShell>);

    expect(screen.getByText(dict.states.unknown)).toBeInTheDocument();
    expect(screen.getByText("—")).toBeInTheDocument();
  });

  it("renders the measured skew once samples arrive", () => {
    resetBenchStore({
      samples: [{ at: Date.now(), skewMs: 1500, roundTripMs: 10 }],
    });
    renderWithLocale(<AppShell>x</AppShell>);

    expect(screen.getByText("+1.50s")).toBeInTheDocument();
  });

  it("reports a device that has gone offline", () => {
    resetBenchStore({
      health: {
        state: "offline",
        consecutiveFailures: 10,
        consecutiveSuccesses: 0,
        successes: 0,
        failures: 10,
        window: new Array(10).fill(false),
        latencies: [],
      },
    });
    const { dict } = renderWithLocale(<AppShell>x</AppShell>);

    expect(screen.getByText(dict.states.offline)).toBeInTheDocument();
  });
});

describe("AppShell engine controls", () => {
  it("starts and stops the poll loop", async () => {
    const { dict } = renderWithLocale(<AppShell>x</AppShell>);

    await userEvent.click(screen.getByRole("button", { name: dict.common.start }));
    expect(useBenchStore.getState().running).toBe(true);

    await userEvent.click(screen.getByRole("button", { name: dict.common.stop }));
    expect(useBenchStore.getState().running).toBe(false);
  });

  it("resets the run", async () => {
    resetBenchStore({
      counters: {
        successfulReads: 7,
        failedReads: 0,
        reconnects: 0,
        retries: 0,
        protocolErrors: 0,
        exhaustedPolls: 0,
        connections: 1,
      },
    });
    const { dict } = renderWithLocale(<AppShell>x</AppShell>);

    await userEvent.click(screen.getByRole("button", { name: dict.common.reset }));

    expect(useBenchStore.getState().counters.successfulReads).toBe(0);
  });

  it("drives the engine on a timer only while running", () => {
    jest.useFakeTimers();
    try {
      renderWithLocale(<AppShell>x</AppShell>);

      act(() => {
        useBenchStore.getState().patchGateway({ scheduleMode: "interval", intervalMs: 5000 });
        useBenchStore.getState().start();
      });
      act(() => {
        jest.advanceTimersByTime(600);
      });

      expect(useBenchStore.getState().counters.successfulReads).toBeGreaterThan(0);

      const reached = useBenchStore.getState().counters.successfulReads;
      act(() => {
        useBenchStore.getState().stop();
        jest.advanceTimersByTime(20_000);
      });

      expect(useBenchStore.getState().counters.successfulReads).toBe(reached);
    } finally {
      jest.useRealTimers();
    }
  });
});
