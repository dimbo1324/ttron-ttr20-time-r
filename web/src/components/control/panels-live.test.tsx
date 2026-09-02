import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { useBenchStore } from "@/stores/bench-store";
import { getDictionary } from "@/i18n";
import { createFormatter } from "@/lib/format";
import { renderWithLocale, resetBenchStore } from "@/test/utils";

import { EmulatorPanel } from "./emulator-panel";
import { GatewayPanel } from "./gateway-panel";

/** Readouts are locale-formatted, so expectations are derived, not hardcoded. */
const format = createFormatter("ru", getDictionary("ru").units);


const NOW = Date.UTC(2026, 8, 2, 12, 0, 0);

/**
 * `userEvent` cannot drag a range input, so the change is dispatched on the
 * element directly — the assertion is about the handler behind the slider.
 */
function slide(element: HTMLElement, value: string) {
  const setter = Object.getOwnPropertyDescriptor(
    window.HTMLInputElement.prototype,
    "value",
  )?.set;
  setter?.call(element, value);
  element.dispatchEvent(new Event("change", { bubbles: true }));
}

beforeEach(() => {
  jest.spyOn(Date, "now").mockReturnValue(NOW);
  resetBenchStore();
});

describe("EmulatorPanel sliders", () => {
  it.each([
    { key: "responseDelay" as const, value: "900", read: () => useBenchStore.getState().faults.responseDelayMs, want: 900 },
    { key: "badChecksum" as const, value: "0.35", read: () => useBenchStore.getState().faults.badChecksumProb, want: 0.35 },
    { key: "fragment" as const, value: "0.25", read: () => useBenchStore.getState().faults.fragmentProb, want: 0.25 },
    { key: "clockOffset" as const, value: "45000", read: () => useBenchStore.getState().faults.clockOffsetMs, want: 45_000 },
    { key: "clockDrift" as const, value: "-48000", read: () => useBenchStore.getState().faults.clockDriftPerDayMs, want: -48_000 },
  ])("$key writes its value to the device", ({ key, value, read, want }) => {
    const { dict } = renderWithLocale(<EmulatorPanel />);

    slide(screen.getByRole("slider", { name: dict.emulator[key] }), value);

    expect(read()).toBe(want);
  });

  it("moves the fragment delay once fragmentation is enabled", () => {
    resetBenchStore({
      faults: { ...useBenchStore.getState().faults, fragmentProb: 0.5 },
    });
    const { dict } = renderWithLocale(<EmulatorPanel />);

    slide(screen.getByRole("slider", { name: dict.emulator.fragmentDelay }), "200");

    expect(useBenchStore.getState().faults.fragmentDelayMs).toBe(200);
  });

  it("renders a zero clock offset as a bare zero", () => {
    const { dict } = renderWithLocale(<EmulatorPanel />);
    const offset = screen.getByRole("slider", { name: dict.emulator.clockOffset });

    expect(offset).toHaveValue("0");
    expect(screen.getAllByText("0").length).toBeGreaterThan(0);
  });

  it("labels a negative drift with its sign", () => {
    resetBenchStore({
      faults: { ...useBenchStore.getState().faults, clockDriftPerDayMs: -48_000 },
    });
    const { dict } = renderWithLocale(<EmulatorPanel />);

    expect(screen.getByText(format.driftPerDay(-48_000))).toBeInTheDocument();
  });

  it("edits the model and the firmware", async () => {
    const { dict } = renderWithLocale(<EmulatorPanel />);

    const model = screen.getByLabelText(dict.overview.model);
    await userEvent.clear(model);
    await userEvent.type(model, "TTR21");

    const firmware = screen.getByLabelText(dict.overview.firmware);
    await userEvent.clear(firmware);
    await userEvent.type(firmware, "2.0.0");

    expect(useBenchStore.getState().identity.model).toBe("TTR21");
    expect(useBenchStore.getState().identity.firmware).toBe("2.0.0");
  });
});

describe("GatewayPanel while polling", () => {
  beforeEach(() => {
    resetBenchStore({ running: true, connected: true, nextPollAt: NOW + 2500 });
  });

  it("reports the loop as running and connected", () => {
    const { dict } = renderWithLocale(<GatewayPanel />);

    expect(screen.getByText(dict.gateway.pollingRunning)).toBeInTheDocument();
    expect(screen.getByText(dict.common.connected)).toBeInTheDocument();
  });

  it("counts down to the next poll", () => {
    renderWithLocale(<GatewayPanel />);

    expect(screen.getByText(format.duration(2500))).toBeInTheDocument();
  });

  it("explains the aligned schedule while it is selected", () => {
    const { dict } = renderWithLocale(<GatewayPanel />);

    expect(screen.getByText(dict.gateway.scheduleAlignedHint)).toBeInTheDocument();
  });

  it("explains the interval schedule once switched", async () => {
    const { dict } = renderWithLocale(<GatewayPanel />);

    await userEvent.click(screen.getByRole("radio", { name: dict.gateway.scheduleInterval }));

    expect(screen.getAllByText(dict.gateway.scheduleIntervalHint).length).toBeGreaterThan(0);
  });

  it("reports a device that has gone offline", () => {
    resetBenchStore({
      running: true,
      connected: false,
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
    const { dict } = renderWithLocale(<GatewayPanel />);

    expect(screen.getByText(dict.common.disconnected)).toBeInTheDocument();
    expect(screen.getByText(dict.states.offline)).toBeInTheDocument();
  });

  it("draws the skew history once there are samples", () => {
    resetBenchStore({
      running: true,
      samples: [0, 500, 1200, 900].map((skewMs, index) => ({
        at: NOW - (4 - index) * 5000,
        skewMs,
        roundTripMs: 10,
      })),
    });
    const { container } = renderWithLocale(<GatewayPanel />);

    // Scoped by radius: the panel's own icons are SVGs carrying circles too.
    expect(container.querySelectorAll('circle[r="1.1"]').length).toBe(4);
  });
});
