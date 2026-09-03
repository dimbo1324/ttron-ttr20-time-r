import { act, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EmulatorPanel } from "@/components/control/emulator-panel";
import { GatewayPanel } from "@/components/control/gateway-panel";
import { Overview } from "@/components/dashboard/overview";
import { ExchangeMonitor } from "@/components/monitor/monitor";
import { api } from "@/lib/api/client";
import { createFormatter } from "@/lib/format";
import { getDictionary } from "@/i18n";
import { useLiveStore } from "@/stores/live-store";
import { useSourceStore } from "@/stores/source-store";
import {
  eventFixture,
  faultModeFixture,
  fleetFixture,
  gatewayStatusFixture,
  historyFixture,
  plain,
  renderWithLocale,
  resetBenchStore,
  resetLiveStore,
  useSource,
} from "@/test/utils";

/**
 * The console pointed at a running stack.
 *
 * These are the tests that hold the line the whole seam exists for: a panel
 * must render live data without knowing it is live, and it must never offer a
 * control that the live source would silently discard.
 */

jest.mock("@/lib/api/client", () => {
  const actual = jest.requireActual<typeof import("@/lib/api/client")>("@/lib/api/client");
  return {
    ...actual,
    api: {
      gatewayStatus: jest.fn(),
      gatewayEvents: jest.fn(),
      gatewayHistory: jest.fn(),
      gatewayFleet: jest.fn(),
      startPolling: jest.fn(),
      stopPolling: jest.fn(),
      faultMode: jest.fn(),
      setFaultMode: jest.fn(),
      emulatorStatus: jest.fn(),
    },
  };
});

const mocked = api as jest.Mocked<typeof api>;
const format = createFormatter("ru", getDictionary("ru").units);

async function goLive(statusOverrides: Record<string, unknown> = {}) {
  mocked.gatewayStatus.mockResolvedValue(gatewayStatusFixture(statusOverrides) as never);
  mocked.gatewayEvents.mockResolvedValue([
    eventFixture({ id: 1, direction: "TX" }),
    eventFixture({ id: 2, direction: "RX", timestamp: "2026-09-02T12:00:00.010Z" }),
    eventFixture({
      id: 3,
      direction: "ERR",
      rawHex: "",
      error: "decode frame: invalid checksum",
      timestamp: "2026-09-02T12:00:00.020Z",
    }),
  ] as never);
  mocked.gatewayHistory.mockResolvedValue(historyFixture() as never);
  mocked.gatewayFleet.mockResolvedValue(fleetFixture() as never);
  mocked.faultMode.mockResolvedValue(faultModeFixture({ responseDelayMs: 250 }) as never);

  useSource("live");
  await act(async () => {
    await useLiveStore.getState().refresh();
    await useLiveStore.getState().loadFaults();
  });
}

/** A panel, found by its heading, so a figure is asserted where it belongs. */
function panel(title: string) {
  return within(screen.getByRole("heading", { name: title }).closest("section")!);
}

beforeEach(() => {
  window.localStorage.clear();
  resetBenchStore();
  resetLiveStore();
  useSourceStore.setState({ source: "bench", hydrated: true });
});

describe("Overview on the live source", () => {
  it("draws the gateway's own measurements", async () => {
    await goLive();
    const { dict } = renderWithLocale(<Overview />);

    // Scoped to their own cards: the same figures appear again in the fleet
    // row below, which is the point rather than a duplication to avoid.
    const skew = panel(dict.overview.skew);
    expect(skew.getByText(plain(format.duration(-2500, { signed: true })))).toBeInTheDocument();
    expect(skew.getByText(dict.states.warn)).toBeInTheDocument();

    const health = panel(dict.gateway.healthPolicy);
    expect(health.getByText(plain(format.percent(0.875)))).toBeInTheDocument();
    expect(health.getByText(dict.states.degraded)).toBeInTheDocument();
  });

  it("names the device it is polling", async () => {
    await goLive();
    renderWithLocale(<Overview />);

    expect(screen.getAllByText("127.0.0.1:9000").length).toBeGreaterThan(0);
  });

  it("shows the identity the probe actually read", async () => {
    await goLive();
    renderWithLocale(<Overview />);

    expect(screen.getByText("SN-42")).toBeInTheDocument();
  });

  it("omits the clock rows the live source cannot measure", async () => {
    await goLive();
    const { dict } = renderWithLocale(<Overview />);

    // A "0" there would claim the device clock had been checked against an
    // injected offset and found correct.
    expect(screen.queryByText(dict.emulator.clockOffset)).not.toBeInTheDocument();
    expect(screen.queryByText(dict.emulator.clockDrift)).not.toBeInTheDocument();
  });

  it("lists the fleet", async () => {
    await goLive();
    const { dict } = renderWithLocale(<Overview />);

    expect(screen.getByRole("heading", { name: dict.fleet.title })).toBeInTheDocument();
    expect(within(screen.getByRole("table")).getByText("TTR20 beta")).toBeInTheDocument();
  });

  it("shows no fleet on the bench, because a simulated device is not one", () => {
    const { dict } = renderWithLocale(<Overview />);

    expect(screen.queryByRole("heading", { name: dict.fleet.title })).not.toBeInTheDocument();
  });
});

describe("GatewayPanel on the live source", () => {
  it("says the settings belong to the gateway process", async () => {
    await goLive();
    const { dict } = renderWithLocale(<GatewayPanel />);

    expect(screen.getByText(dict.source.readOnlyHint)).toBeInTheDocument();
    expect(screen.getByText(dict.source.readOnly)).toBeInTheDocument();
  });

  it("offers no control that would be discarded", async () => {
    await goLive();
    renderWithLocale(<GatewayPanel />);

    // Not a disabled <select>: one can only display a value that happens to
    // be among its options, and a gateway set to an interval this UI never
    // offers would render blank.
    expect(screen.queryAllByRole("combobox")).toHaveLength(0);
  });

  it("shows the gateway's actual configuration", async () => {
    await goLive();
    const { dict } = renderWithLocale(<GatewayPanel />);

    const value = (label: string) => screen.getByLabelText(label).textContent ?? "";
    expect(plain(value(dict.gateway.interval))).toBe(plain(format.duration(60_000)));
    expect(plain(value(dict.gateway.offset))).toBe(plain(format.duration(5000)));
    expect(plain(value(dict.gateway.requestTimeout))).toBe(plain(format.duration(1500)));
    expect(value(dict.gateway.retryAttempts)).toBe("2");
    expect(plain(value(dict.gateway.warnThreshold))).toBe(plain(format.duration(2000)));
    expect(value(dict.gateway.degradeAfter)).toBe("3");
  });

  it("keeps the schedule mode readable without a control", async () => {
    await goLive();
    const { dict } = renderWithLocale(<GatewayPanel />);

    expect(screen.getByLabelText(dict.gateway.scheduleMode)).toHaveTextContent(
      dict.gateway.scheduleAligned,
    );
  });

  it("explains that the counters are the gateway's", async () => {
    await goLive();
    const { dict } = renderWithLocale(<GatewayPanel />);

    expect(screen.getByText(dict.source.noReset)).toBeInTheDocument();
  });

  it("is fully editable on the bench", () => {
    const { dict } = renderWithLocale(<GatewayPanel />);

    expect(screen.queryByText(dict.source.readOnly)).not.toBeInTheDocument();
    expect(screen.getAllByRole("combobox").length).toBeGreaterThan(0);
  });
});

describe("EmulatorPanel on the live source", () => {
  it("says the switches reach the real emulator", async () => {
    await goLive();
    const { dict } = renderWithLocale(<EmulatorPanel />);

    expect(screen.getByText(dict.source.liveFaultsHint)).toBeInTheDocument();
  });

  it("shows the emulator's current faults", async () => {
    await goLive();
    const { dict } = renderWithLocale(<EmulatorPanel />);

    expect(screen.getByLabelText(dict.emulator.responseDelay)).toHaveValue("250");
  });

  it("sends a change to the emulator", async () => {
    await goLive();
    mocked.setFaultMode.mockResolvedValue(faultModeFixture({ noResponse: true }) as never);
    const { dict } = renderWithLocale(<EmulatorPanel />);

    await userEvent.click(screen.getByLabelText(dict.emulator.noResponse));

    expect(mocked.setFaultMode).toHaveBeenCalledWith(expect.objectContaining({ noResponse: true }));
  });

  it("disables the clock sliders and says why", async () => {
    await goLive();
    const { dict } = renderWithLocale(<EmulatorPanel />);

    expect(screen.getByLabelText(dict.emulator.clockOffset)).toBeDisabled();
    expect(screen.getByLabelText(dict.emulator.clockDrift)).toBeDisabled();
    expect(screen.getByText(dict.source.benchOnly)).toBeInTheDocument();
    expect(screen.getByText(dict.source.benchOnlyHint)).toBeInTheDocument();
  });

  it("hides the preset that only moves the clock", async () => {
    await goLive();
    const { dict } = renderWithLocale(<EmulatorPanel />);

    // A preset that would clear the other faults and change nothing else is
    // a trap, not a shortcut.
    expect(
      screen.queryByRole("button", { name: dict.emulator.presetDriftingClock }),
    ).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: dict.emulator.presetNoisyLine })).toBeInTheDocument();
  });

  it("shows the nameplate read-only, because the emulator owns it", async () => {
    await goLive();
    const { dict } = renderWithLocale(<EmulatorPanel />);

    expect(screen.getByText(dict.source.identityLive)).toBeInTheDocument();
    expect(screen.getByLabelText(dict.overview.model)).toHaveTextContent("TTR20");
  });

  it("says so when the emulator is not answering", async () => {
    await goLive();
    act(() => {
      useLiveStore.setState({ faults: null });
    });
    const { dict } = renderWithLocale(<EmulatorPanel />);

    expect(screen.getByText(dict.source.emulatorUnavailable)).toBeInTheDocument();
    expect(screen.getByLabelText(dict.emulator.responseDelay)).toBeDisabled();
  });

  it("keeps the clock sliders on the bench", () => {
    const { dict } = renderWithLocale(<EmulatorPanel />);

    expect(screen.getByLabelText(dict.emulator.clockOffset)).toBeEnabled();
    expect(screen.queryByText(dict.source.benchOnly)).not.toBeInTheDocument();
  });
});

describe("ExchangeMonitor on the live source", () => {
  it("shows the gateway's own frames", async () => {
    await goLive();
    renderWithLocale(<ExchangeMonitor />);

    expect(screen.getAllByText("68 03 68 00 01 01 02 16").length).toBe(2);
  });

  it("translates an error it recognises in the gateway's wording", async () => {
    await goLive();
    const { dict } = renderWithLocale(<ExchangeMonitor />);

    expect(screen.getByText(dict.events.errors.invalidChecksum)).toBeInTheDocument();
  });

  it("keeps the gateway's own words in the detail panel", async () => {
    await goLive();
    const { dict } = renderWithLocale(<ExchangeMonitor />);

    await userEvent.click(screen.getByText(dict.events.errors.invalidChecksum));

    // The message names the sentinel that produced it, which is what someone
    // reading the Go log needs to match against.
    expect(screen.getByText("decode frame: invalid checksum")).toBeInTheDocument();
  });

  it("offers no way to erase the gateway's history", async () => {
    await goLive();
    const { dict } = renderWithLocale(<ExchangeMonitor />);

    expect(screen.queryByRole("button", { name: dict.monitor.clearLog })).not.toBeInTheDocument();
  });

  it("still clears the bench's own log", () => {
    const { dict } = renderWithLocale(<ExchangeMonitor />);

    expect(screen.getByRole("button", { name: dict.monitor.clearLog })).toBeInTheDocument();
  });

  it("shows the gateway's counters", async () => {
    await goLive();
    const { dict } = renderWithLocale(<ExchangeMonitor />);

    const counters = screen.getByText(dict.monitor.counters).closest("section")!;
    expect(within(counters).getByText("42")).toBeInTheDocument();
    expect(within(counters).getByText("7")).toBeInTheDocument();
  });
});
