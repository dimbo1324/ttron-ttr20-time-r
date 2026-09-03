import { act, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EmulatorPanel } from "@/components/control/emulator-panel";
import { GatewayPanel } from "@/components/control/gateway-panel";
import { Overview } from "@/components/dashboard/overview";
import { ExchangeMonitor } from "@/components/monitor/monitor";
import { api, ApiError } from "@/lib/api/client";
import { createFormatter } from "@/lib/format";
import { getDictionary } from "@/i18n";
import { useBenchStore } from "@/stores/bench-store";
import { useLiveStore } from "@/stores/live-store";
import { useSourceStore } from "@/stores/source-store";
import {
  eventFixture,
  faultModeFixture,
  fleetFixture,
  gatewayStatusFixture,
  historyFixture,
  plain,
  settingsFixture,
  renderWithLocale,
  resetBenchStore,
  resetLiveStore,
  selectSource,
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
      updateSettings: jest.fn(),
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

  selectSource("live");
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
  it("says the changes reach the running gateway, and which device", async () => {
    await goLive();
    const { dict } = renderWithLocale(<GatewayPanel />);

    expect(screen.getByText(dict.source.liveSettingsHint)).toBeInTheDocument();
    // In inventory mode the control plane is bound to the primary device, so a
    // change here reconfigures that one and not the fleet beside it.
    expect(screen.getByText(dict.source.controlledDevice, { exact: false })).toBeInTheDocument();
    expect(screen.getByText("TTR20 alpha")).toBeInTheDocument();
  });

  it("shows the gateway's actual configuration", async () => {
    await goLive();
    const { dict } = renderWithLocale(<GatewayPanel />);

    const value = (label: string) =>
      (screen.getByLabelText(label) as HTMLSelectElement).value;
    expect(value(dict.gateway.interval)).toBe("60000");
    expect(value(dict.gateway.offset)).toBe("5000");
    expect(value(dict.gateway.requestTimeout)).toBe("1500");
    expect(value(dict.gateway.retryAttempts)).toBe("2");
    expect(value(dict.gateway.retryDelay)).toBe("200");
    expect(value(dict.gateway.warnThreshold)).toBe("2000");
    expect(value(dict.gateway.degradeAfter)).toBe("3");
  });

  it("sends a changed interval to the gateway, with the rest unchanged", async () => {
    await goLive();
    mocked.updateSettings.mockResolvedValue({
      settings: settingsFixture({ pollIntervalMs: 10_000 }),
      status: gatewayStatusFixture(),
    } as never);
    const { dict } = renderWithLocale(<GatewayPanel />);

    await userEvent.selectOptions(screen.getByLabelText(dict.gateway.interval), "10000");

    // The whole configuration goes over, filled in from what the gateway is
    // running now rather than from a snapshot taken when the page loaded.
    expect(mocked.updateSettings).toHaveBeenCalledWith({
      scheduleMode: "aligned",
      pollIntervalMs: 10_000,
      pollOffsetMs: 5000,
      requestTimeoutMs: 1500,
      retryAttempts: 2,
      retryDelayMs: 200,
      clockWarnMs: 2000,
      clockCriticalMs: 30_000,
      degradeAfter: 3,
      offlineAfter: 10,
      recoverAfter: 2,
    });
  });

  it("drops the offset when the schedule stops being aligned", async () => {
    await goLive();
    mocked.updateSettings.mockResolvedValue({
      settings: settingsFixture({ scheduleMode: "interval", pollOffsetMs: 0 }),
      status: gatewayStatusFixture(),
    } as never);
    const { dict } = renderWithLocale(<GatewayPanel />);

    await userEvent.click(screen.getByRole("radio", { name: dict.gateway.scheduleInterval }));

    // An offset only means something to an aligned schedule, and the gateway
    // rejects one that is not below the interval.
    expect(mocked.updateSettings).toHaveBeenCalledWith(
      expect.objectContaining({ scheduleMode: "interval", pollOffsetMs: 0 }),
    );
  });

  it("keeps the offset locked while the schedule is by interval", async () => {
    await goLive({ schedule: { ...gatewayStatusFixture().schedule, mode: "interval", offsetMs: 0 } });
    const { dict } = renderWithLocale(<GatewayPanel />);

    expect(screen.getByLabelText(dict.gateway.offset).tagName).toBe("OUTPUT");
  });

  it("offers the gateway's own value even when this UI never suggests it", async () => {
    await goLive({
      schedule: { ...gatewayStatusFixture().schedule, intervalMs: 7000 },
      requestTimeoutMs: 900,
    });
    const { dict } = renderWithLocale(<GatewayPanel />);

    // Without this the select renders blank, which reads as "not set", and the
    // first edit silently snaps the value to whichever option came first.
    const interval = screen.getByLabelText(dict.gateway.interval) as HTMLSelectElement;
    expect(interval.value).toBe("7000");
    expect(within(interval).getByRole("option", { name: /7/ })).toBeInTheDocument();
  });

  it("never offers a value the gateway would refuse", async () => {
    await goLive({ schedule: { ...gatewayStatusFixture().schedule, intervalMs: 1000 } });
    const { dict } = renderWithLocale(<GatewayPanel />);

    // A request that may outlast its own interval is not a schedule.
    const timeout = screen.getByLabelText(dict.gateway.requestTimeout);
    const offered = within(timeout)
      .getAllByRole("option")
      .map((option) => Number((option as HTMLOptionElement).value));
    expect(offered.filter((value) => value >= 1000 && value !== 1500)).toHaveLength(0);
  });

  it("freezes every control while a write is in flight", async () => {
    await goLive();
    act(() => {
      useLiveStore.setState({ busy: true });
    });
    const { dict } = renderWithLocale(<GatewayPanel />);

    // The gateway takes the whole configuration at once, so a second edit sent
    // mid-flight would race the first.
    expect(screen.getByLabelText(dict.gateway.interval).tagName).toBe("OUTPUT");
    expect(screen.getByLabelText(dict.gateway.warnThreshold).tagName).toBe("OUTPUT");
  });

  it("shows values rather than knobs while the link is down", async () => {
    await goLive();
    act(() => {
      useLiveStore.setState({ link: "unreachable", status: null });
    });
    const { dict } = renderWithLocale(<GatewayPanel />);

    expect(screen.getByText(dict.source.settingsUnavailable)).toBeInTheDocument();
    expect(screen.getByText(dict.source.readOnly)).toBeInTheDocument();
    expect(screen.queryAllByRole("combobox")).toHaveLength(0);
  });

  it("keeps a refusal on screen instead of letting the next poll wipe it", async () => {
    await goLive();
    mocked.updateSettings.mockRejectedValue(
      new ApiError(
        "invalid gateway settings: request timeout must be below the poll interval",
        400,
        "INVALID_ARGUMENT",
      ),
    );
    const { dict } = renderWithLocale(<GatewayPanel />);

    await userEvent.selectOptions(
      screen.getByLabelText(dict.gateway.warnThreshold),
      "1000",
    );

    // The link-level error is rewritten by every refresh a second apart; a
    // rejection put there would flash and vanish before it could be read.
    await waitFor(() =>
      expect(screen.getByText(dict.source.settingsRejected)).toBeInTheDocument(),
    );
    expect(screen.getByText(/below the poll interval/)).toBeInTheDocument();

    act(() => {
      void useLiveStore.getState().refresh();
    });
    await waitFor(() =>
      expect(screen.getByText(dict.source.settingsRejected)).toBeInTheDocument(),
    );
  });

  it("says nothing about a refusal when there has not been one", async () => {
    await goLive();
    const { dict } = renderWithLocale(<GatewayPanel />);

    expect(screen.queryByText(dict.source.settingsRejected)).not.toBeInTheDocument();
  });

  it("explains that the counters are the gateway's", async () => {
    await goLive();
    const { dict } = renderWithLocale(<GatewayPanel />);

    expect(screen.getByText(dict.source.noReset)).toBeInTheDocument();
  });

  it("is editable on the bench, and says nothing about a device", () => {
    const { dict } = renderWithLocale(<GatewayPanel />);

    expect(screen.getAllByRole("combobox").length).toBeGreaterThan(0);
    expect(screen.queryByText(dict.source.liveSettingsHint)).not.toBeInTheDocument();
    expect(screen.queryByText(dict.source.controlledDevice, { exact: false })).not.toBeInTheDocument();
  });

  it("writes to the bench engine rather than to the API", async () => {
    const { dict } = renderWithLocale(<GatewayPanel />);

    await userEvent.selectOptions(screen.getByLabelText(dict.gateway.interval), "10000");

    expect(useBenchStore.getState().gateway.intervalMs).toBe(10_000);
    expect(mocked.updateSettings).not.toHaveBeenCalled();
  });

  it("changes the retry delay the engine actually backs off by", async () => {
    const { dict } = renderWithLocale(<GatewayPanel />);

    await userEvent.selectOptions(screen.getByLabelText(dict.gateway.retryDelay), "500");

    // The engine doubles from this on each further attempt, the same shape the
    // Go retry policy uses -- a bench backing off on a schedule of its own
    // would make a retry storm look different here than on the wire.
    expect(useBenchStore.getState().gateway.retryDelayMs).toBe(500);
  });

  it("merges a threshold change into the bench engine's nested settings", async () => {
    const { dict } = renderWithLocale(<GatewayPanel />);

    await userEvent.selectOptions(screen.getByLabelText(dict.gateway.warnThreshold), "1000");

    const thresholds = useBenchStore.getState().gateway.thresholds;
    // The engine keeps these nested, so a partial update has to be merged
    // against what is there rather than assigned over it.
    expect(thresholds.warnMs).toBe(1000);
    expect(thresholds.criticalMs).toBe(30_000);
  });

  it("merges a policy change the same way", async () => {
    const { dict } = renderWithLocale(<GatewayPanel />);

    await userEvent.selectOptions(screen.getByLabelText(dict.gateway.recoverAfter), "3");

    const policy = useBenchStore.getState().gateway.policy;
    expect(policy.recoverAfter).toBe(3);
    expect(policy.degradeAfter).toBe(3);
    expect(policy.offlineAfter).toBe(10);
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
