import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { Overview } from "@/components/dashboard/overview";
import { AppShell } from "@/components/layout/app-shell";
import { ExchangeMonitor } from "@/components/monitor/monitor";
import { FrameAnalyzer } from "@/components/protocol/analyzer";
import { FIELD_STYLE } from "@/components/protocol/frame-view";
import { badgeVariants } from "@/components/ui/badge";
import { buttonVariants } from "@/components/ui/button";
import { Field, Input } from "@/components/ui/controls";
import { StatusDot } from "@/components/ui/status-dot";
import {
  defaultLocale,
  getLocaleFromPathname,
  isLocale,
  LOCALE_COOKIE,
  LOCALE_COOKIE_MAX_AGE,
  LOCALE_NAMES,
  LOCALE_SHORT,
  locales,
  withLocale,
} from "@/i18n";
import { renderWithLocale, resetBenchStore } from "@/test/utils";

/**
 * The states a screen only reaches once something has gone wrong, plus the
 * exports that carry no behaviour of their own. Both are cheap to leave
 * untested and both are what a reader reaches for when they are already
 * confused, so they get the same treatment as the happy path.
 */

const NOW = Date.UTC(2026, 8, 2, 12, 0, 0);

let pathname = "/ru";
jest.mock("next/navigation", () => ({
  usePathname: () => pathname,
}));

beforeEach(() => {
  pathname = "/ru";
  jest.spyOn(Date, "now").mockReturnValue(NOW);
  resetBenchStore();
});

describe("re-exported locale surface", () => {
  it("is reachable from the package entry", () => {
    expect(locales).toContain(defaultLocale);
    expect(isLocale(defaultLocale)).toBe(true);
    expect(getLocaleFromPathname("/en/monitor")).toBe("en");
    expect(withLocale("/ru/monitor", "en")).toBe("/en/monitor");
    expect(LOCALE_NAMES.ru).toBeTruthy();
    expect(LOCALE_SHORT.en).toBe("EN");
    expect(LOCALE_COOKIE).toBeTruthy();
    expect(LOCALE_COOKIE_MAX_AGE).toBeGreaterThan(0);
  });
});

describe("variant helpers", () => {
  it("compose classes for callers outside the components", () => {
    expect(buttonVariants({ variant: "outline", size: "sm" })).toContain("border");
    expect(badgeVariants({ tone: "danger" })).toContain("destructive");
    expect(FIELD_STYLE.checksum.text).toBe("text-field-checksum");
  });
});

describe("StatusDot default", () => {
  it("falls back to the unknown tone", () => {
    const { container } = render(<StatusDot />);

    expect(container.firstChild).toHaveClass("text-faint-foreground");
  });
});

describe("Field with a caller-supplied id", () => {
  it("uses the id it was given rather than minting one", () => {
    render(
      <Field label="Адрес" htmlFor="explicit-id">
        <Input id="explicit-id" defaultValue="1" />
      </Field>,
    );

    expect(screen.getByLabelText("Адрес")).toHaveAttribute("id", "explicit-id");
  });

  it("tolerates a child that is not a single element", () => {
    render(
      <Field label="Пара">
        <Input aria-label="a" />
        <Input aria-label="b" />
      </Field>,
    );

    expect(screen.getByLabelText("a")).toBeInTheDocument();
    expect(screen.getByLabelText("b")).toBeInTheDocument();
  });
});

describe("AppShell clock tones", () => {
  it.each([
    { name: "warning", skewMs: 5000 },
    { name: "critical", skewMs: 45_000 },
  ])("renders the $name reading", ({ skewMs }) => {
    resetBenchStore({
      samples: new Array(4).fill(null).map((_, index) => ({
        at: NOW - (4 - index) * 60_000,
        skewMs,
        roundTripMs: 10,
      })),
    });
    renderWithLocale(<AppShell>x</AppShell>);

    expect(screen.getByText(`+${(skewMs / 1000).toFixed(2)}s`)).toBeInTheDocument();
  });

  it("renders a device stuck in a degraded state", () => {
    resetBenchStore({
      health: {
        state: "degraded",
        consecutiveFailures: 3,
        consecutiveSuccesses: 0,
        successes: 5,
        failures: 3,
        window: [true, true, false, false, false],
        latencies: [10],
      },
    });
    const { dict } = renderWithLocale(<AppShell>x</AppShell>);

    expect(screen.getByText(dict.states.degraded)).toBeInTheDocument();
  });
});

describe("Overview without a drift line", () => {
  it("leaves the drift blank until the samples support one", () => {
    resetBenchStore({
      running: true,
      samples: [{ at: NOW, skewMs: 100, roundTripMs: 10 }],
    });
    const { dict } = renderWithLocale(<Overview />);

    expect(screen.getByText(dict.overview.driftHint)).toBeInTheDocument();
  });

  it("omits the identity block until the probe has run", () => {
    resetBenchStore({ running: true });
    const { dict } = renderWithLocale(<Overview />);

    expect(screen.queryByText(dict.overview.model)).not.toBeInTheDocument();
  });

  it("shows the interval schedule when that is what is configured", () => {
    resetBenchStore({
      gateway: { ...resetGateway(), scheduleMode: "interval" },
    });
    const { dict } = renderWithLocale(<Overview />);

    expect(screen.getByText(dict.gateway.scheduleInterval)).toBeInTheDocument();
  });

  it("renders the emulator clock faults it was given", () => {
    resetBenchStore({
      faults: {
        responseDelayMs: 250,
        badChecksumProb: 0,
        fragmentProb: 0,
        fragmentDelayMs: 40,
        noResponse: false,
        closeAfterRequest: false,
        clockOffsetMs: -4000,
        clockDriftPerDayMs: 24_000,
      },
    });
    renderWithLocale(<Overview />);

    expect(screen.getByText("-4.00 s")).toBeInTheDocument();
    expect(screen.getByText("250 ms")).toBeInTheDocument();
  });
});

describe("ExchangeMonitor unusual rows", () => {
  it("renders a system note that carries no transition", () => {
    resetBenchStore({
      events: [
        {
          id: 1,
          at: NOW,
          direction: "sys",
          source: "gateway",
          command: "connect",
          bytes: [],
          note: "pollingStarted",
          cycle: 0,
          attempt: 0,
        },
      ],
    });
    const { dict } = renderWithLocale(<ExchangeMonitor />);

    expect(screen.getAllByText(dict.gateway.pollingRunning).length).toBeGreaterThan(0);
  });

  it("falls back to the raw key for a note it cannot translate", () => {
    resetBenchStore({
      events: [
        {
          id: 1,
          at: NOW,
          direction: "sys",
          source: "gateway",
          command: "custom",
          bytes: [],
          note: "somethingNew",
          cycle: 0,
          attempt: 0,
        },
      ],
    });
    renderWithLocale(<ExchangeMonitor />);

    expect(screen.getByText("somethingNew")).toBeInTheDocument();
  });

  it("renders a row that carries neither frame nor note", () => {
    resetBenchStore({
      events: [
        {
          id: 1,
          at: NOW,
          direction: "sys",
          source: "gateway",
          command: "silent",
          bytes: [],
          cycle: 0,
          attempt: 0,
        },
      ],
    });
    renderWithLocale(<ExchangeMonitor />);

    expect(screen.getAllByText("silent").length).toBeGreaterThan(0);
  });

  it("shows a dash for a selected event with nothing to say", async () => {
    resetBenchStore({
      events: [
        {
          id: 1,
          at: NOW,
          direction: "sys",
          source: "gateway",
          command: "silent",
          bytes: [],
          cycle: 0,
          attempt: 0,
        },
      ],
    });
    renderWithLocale(<ExchangeMonitor />);

    await userEvent.click(screen.getAllByText("silent")[0]!);

    expect(screen.getByText("—")).toBeInTheDocument();
  });
});

describe("FrameAnalyzer builder for identity", () => {
  it("builds an identity request when the direction is a request", async () => {
    const { dict } = renderWithLocale(<FrameAnalyzer />);

    await userEvent.selectOptions(screen.getByLabelText(dict.protocol.command), "2");
    await userEvent.click(screen.getByRole("button", { name: dict.protocol.insert }));

    expect(screen.getAllByText(dict.common.request).length).toBeGreaterThanOrEqual(2);
    expect(screen.getByText("read-identity")).toBeInTheDocument();
  });

  it("marks a broken frame inside a scanned stream", async () => {
    const { dict } = renderWithLocale(<FrameAnalyzer />);
    const input = screen.getByLabelText(dict.protocol.input);

    await userEvent.clear(input);
    await userEvent.paste("68 03 68 00 01 01 FF 16 68 03 68 00 01 01 02 16");

    expect(screen.getAllByText(dict.protocol.invalid).length).toBeGreaterThan(0);
    expect(screen.getAllByText(dict.protocol.valid).length).toBeGreaterThan(0);
  });
});

/** The store's shipped gateway defaults, for tests that only override one field. */
function resetGateway() {
  return {
    scheduleMode: "aligned" as const,
    intervalMs: 5000,
    offsetMs: 0,
    requestTimeoutMs: 1500,
    retryAttempts: 2,
    thresholds: { warnMs: 2000, criticalMs: 30_000 },
    policy: { degradeAfter: 3, offlineAfter: 10, recoverAfter: 2 },
    identityProbe: false,
  };
}
