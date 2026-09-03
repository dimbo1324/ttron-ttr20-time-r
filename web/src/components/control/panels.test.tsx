import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { formatDeviceTime } from "@/lib/ft12";
import { useBenchStore } from "@/stores/bench-store";
import { getDictionary } from "@/i18n";
import { createFormatter } from "@/lib/format";
import { plain, renderWithLocale, resetBenchStore } from "@/test/utils";

import { EmulatorPanel } from "./emulator-panel";
import { GatewayPanel } from "./gateway-panel";

/** Readouts are locale-formatted, so expectations are derived, not hardcoded. */
const format = createFormatter("ru", getDictionary("ru").units);


const NOW = Date.UTC(2026, 8, 2, 12, 0, 0);

beforeEach(() => {
  jest.spyOn(Date, "now").mockReturnValue(NOW);
  resetBenchStore();
});

describe("EmulatorPanel faults", () => {
  it("reports a fault-free device", () => {
    const { dict } = renderWithLocale(<EmulatorPanel />);

    expect(screen.getAllByText(dict.emulator.presetHealthy).length).toBeGreaterThan(0);
  });

  it("switches silence on and off", async () => {
    const { dict } = renderWithLocale(<EmulatorPanel />);

    await userEvent.click(screen.getByRole("switch", { name: dict.emulator.noResponse }));
    expect(useBenchStore.getState().faults.noResponse).toBe(true);

    await userEvent.click(screen.getByRole("switch", { name: dict.emulator.noResponse }));
    expect(useBenchStore.getState().faults.noResponse).toBe(false);
  });

  it("switches the connection drop", async () => {
    const { dict } = renderWithLocale(<EmulatorPanel />);

    await userEvent.click(screen.getByRole("switch", { name: dict.emulator.closeAfterRequest }));

    expect(useBenchStore.getState().faults.closeAfterRequest).toBe(true);
  });

  it("counts the faults an operator has switched on", async () => {
    const { dict } = renderWithLocale(<EmulatorPanel />);

    await userEvent.click(screen.getByRole("switch", { name: dict.emulator.noResponse }));

    expect(screen.getAllByText(`${dict.emulator.activeFaults}: 1`).length).toBeGreaterThan(0);
  });

  it("disables the fragment delay until fragmentation is on", () => {
    const { dict } = renderWithLocale(<EmulatorPanel />);

    expect(screen.getByRole("slider", { name: dict.emulator.fragmentDelay })).toBeDisabled();
  });

  it("enables the fragment delay once fragmentation has a probability", () => {
    resetBenchStore({
      faults: { ...useBenchStore.getState().faults, fragmentProb: 0.5 },
    });
    const { dict } = renderWithLocale(<EmulatorPanel />);

    expect(screen.getByRole("slider", { name: dict.emulator.fragmentDelay })).toBeEnabled();
  });

  it("renders the current fault values", () => {
    resetBenchStore({
      faults: {
        responseDelayMs: 900,
        badChecksumProb: 0.35,
        fragmentProb: 0.25,
        fragmentDelayMs: 60,
        noResponse: false,
        closeAfterRequest: false,
        clockOffsetMs: 4000,
        clockDriftPerDayMs: 45_000,
      },
    });
    renderWithLocale(<EmulatorPanel />);

    expect(screen.getByText(format.duration(900))).toBeInTheDocument();
    expect(screen.getByText(plain(format.percent(0.35, 0)))).toBeInTheDocument();
    expect(screen.getByText(plain(format.percent(0.25, 0)))).toBeInTheDocument();
    expect(screen.getByText(format.duration(4000, { signed: true }))).toBeInTheDocument();
  });
});

describe("EmulatorPanel presets", () => {
  it.each([
    { key: "presetNoisyLine" as const, check: () => useBenchStore.getState().faults.badChecksumProb },
    { key: "presetSlowDevice" as const, check: () => useBenchStore.getState().faults.responseDelayMs },
    { key: "presetDriftingClock" as const, check: () => useBenchStore.getState().faults.clockOffsetMs },
  ])("$key switches faults on", async ({ key, check }) => {
    const { dict } = renderWithLocale(<EmulatorPanel />);

    await userEvent.click(screen.getByRole("button", { name: dict.emulator[key] }));

    expect(check()).toBeGreaterThan(0);
  });

  it("the dead-device preset silences the device", async () => {
    const { dict } = renderWithLocale(<EmulatorPanel />);

    await userEvent.click(screen.getByRole("button", { name: dict.emulator.presetDeadDevice }));

    expect(useBenchStore.getState().faults.noResponse).toBe(true);
  });

  it("the healthy preset clears everything", async () => {
    const { dict } = renderWithLocale(<EmulatorPanel />);

    await userEvent.click(screen.getByRole("button", { name: dict.emulator.presetNoisyLine }));
    await userEvent.click(screen.getAllByRole("button", { name: dict.emulator.presetHealthy })[0]!);

    expect(useBenchStore.getState().faults.badChecksumProb).toBe(0);
    expect(useBenchStore.getState().faults.noResponse).toBe(false);
  });
});

describe("EmulatorPanel device settings", () => {
  it("edits the identity", async () => {
    const { dict } = renderWithLocale(<EmulatorPanel />);
    const serial = screen.getByLabelText(dict.overview.serial);

    await userEvent.clear(serial);
    await userEvent.type(serial, "SN-999");

    expect(useBenchStore.getState().identity.serial).toBe("SN-999");
  });

  it("switches the checksum mode", async () => {
    renderWithLocale(<EmulatorPanel />);

    await userEvent.click(screen.getByRole("radio", { name: "crc16" }));

    expect(useBenchStore.getState().checksumMode).toBe("crc16");
  });

  it("clamps the adapter address to a byte", async () => {
    const { dict } = renderWithLocale(<EmulatorPanel />);
    const address = screen.getByLabelText(dict.protocol.address);

    await userEvent.clear(address);
    await userEvent.type(address, "300");

    expect(useBenchStore.getState().adapterAddress).toBeLessThanOrEqual(255);
  });

  it("previews the response the device would send", () => {
    const { dict } = renderWithLocale(<EmulatorPanel />);

    expect(screen.getAllByText(dict.protocol.valid).length).toBeGreaterThan(0);
    // The wire timestamp is a naive local wall clock, so the expectation is
    // derived the same way the frame was rather than hardcoded to one zone.
    expect(screen.getByText(formatDeviceTime(new Date(NOW)))).toBeInTheDocument();
  });
});

describe("GatewayPanel schedule", () => {
  it("reports the loop as stopped", () => {
    const { dict } = renderWithLocale(<GatewayPanel />);

    expect(screen.getByText(dict.gateway.pollingStopped)).toBeInTheDocument();
  });

  it("switches between interval and aligned mode", async () => {
    const { dict } = renderWithLocale(<GatewayPanel />);

    await userEvent.click(screen.getByRole("radio", { name: dict.gateway.scheduleInterval }));
    expect(useBenchStore.getState().gateway.scheduleMode).toBe("interval");

    await userEvent.click(screen.getByRole("radio", { name: dict.gateway.scheduleAligned }));
    expect(useBenchStore.getState().gateway.scheduleMode).toBe("aligned");
  });

  it("changes the poll interval", async () => {
    const { dict } = renderWithLocale(<GatewayPanel />);

    await userEvent.selectOptions(screen.getByLabelText(dict.gateway.interval), "60000");

    expect(useBenchStore.getState().gateway.intervalMs).toBe(60_000);
  });

  it("offers only offsets that fit inside the interval", async () => {
    const { dict } = renderWithLocale(<GatewayPanel />);
    await userEvent.selectOptions(screen.getByLabelText(dict.gateway.interval), "5000");

    const options = [...screen.getByLabelText(dict.gateway.offset).querySelectorAll("option")];

    expect(options.every((option) => Number(option.value) < 5000)).toBe(true);
  });

  it("stops offering an offset in interval mode", async () => {
    const { dict } = renderWithLocale(<GatewayPanel />);

    await userEvent.click(screen.getByRole("radio", { name: dict.gateway.scheduleInterval }));

    // An offset only means something to an aligned schedule. Shown as its
    // value rather than as a disabled dropdown, which invites a click that
    // will not do anything.
    expect(screen.getByLabelText(dict.gateway.offset).tagName).toBe("OUTPUT");
    expect(screen.queryByRole("combobox", { name: dict.gateway.offset })).not.toBeInTheDocument();
  });

  it("sets the aligned offset", async () => {
    const { dict } = renderWithLocale(<GatewayPanel />);

    await userEvent.selectOptions(screen.getByLabelText(dict.gateway.interval), "60000");
    await userEvent.selectOptions(screen.getByLabelText(dict.gateway.offset), "5000");

    expect(useBenchStore.getState().gateway.offsetMs).toBe(5000);
  });
});

describe("GatewayPanel retries and thresholds", () => {
  it("changes the request timeout and the retry count", async () => {
    const { dict } = renderWithLocale(<GatewayPanel />);

    await userEvent.selectOptions(screen.getByLabelText(dict.gateway.requestTimeout), "3000");
    await userEvent.selectOptions(
      screen.getAllByLabelText(dict.gateway.retryAttempts)[0]!,
      "5",
    );

    expect(useBenchStore.getState().gateway.requestTimeoutMs).toBe(3000);
    expect(useBenchStore.getState().gateway.retryAttempts).toBe(5);
  });

  it("changes the clock thresholds", async () => {
    const { dict } = renderWithLocale(<GatewayPanel />);

    await userEvent.selectOptions(screen.getByLabelText(dict.gateway.warnThreshold), "5000");
    await userEvent.selectOptions(screen.getByLabelText(dict.gateway.criticalThreshold), "60000");

    expect(useBenchStore.getState().gateway.thresholds).toEqual({
      warnMs: 5000,
      criticalMs: 60_000,
    });
  });

  it("changes the health policy", async () => {
    const { dict } = renderWithLocale(<GatewayPanel />);

    await userEvent.selectOptions(screen.getByLabelText(dict.gateway.degradeAfter), "5");
    await userEvent.selectOptions(screen.getByLabelText(dict.gateway.offlineAfter), "15");
    await userEvent.selectOptions(screen.getByLabelText(dict.gateway.recoverAfter), "3");

    expect(useBenchStore.getState().gateway.policy).toEqual({
      degradeAfter: 5,
      offlineAfter: 15,
      recoverAfter: 3,
    });
  });

  it("shows the counters of the current run", () => {
    resetBenchStore({
      counters: {
        successfulReads: 9,
        failedReads: 2,
        reconnects: 1,
        retries: 4,
        protocolErrors: 3,
        exhaustedPolls: 2,
        connections: 5,
      },
    });
    renderWithLocale(<GatewayPanel />);

    expect(screen.getByText("9")).toBeInTheDocument();
    expect(screen.getByText("4")).toBeInTheDocument();
  });

  it("draws a tick for each poll that opened a cycle", () => {
    resetBenchStore({
      events: [0, 5000, 10_000].map((offset, index) => ({
        id: index + 1,
        at: NOW - offset,
        direction: "tx" as const,
        source: "gateway" as const,
        command: "read-time",
        bytes: [],
        cycle: index,
        attempt: 0,
      })),
    });
    const { container } = renderWithLocale(<GatewayPanel />);

    expect(container.querySelectorAll("div[style*='left']").length).toBeGreaterThanOrEqual(3);
  });
});
