import { act, render, renderHook, waitFor } from "@testing-library/react";

import { api } from "@/lib/api/client";
import { useBenchStore } from "@/stores/bench-store";
import { useLiveStore } from "@/stores/live-store";
import { useSourceStore } from "@/stores/source-store";
import {
  faultModeFixture,
  gatewayStatusFixture,
  resetBenchStore,
  resetLiveStore,
  useSource,
} from "@/test/utils";

import {
  useFaultControls,
  useSource as useSourceValue,
  useTelemetry,
  useTelemetryControls,
  useTelemetryEngine,
} from "./use-telemetry";

/**
 * The seam itself: which source a component gets, and which loop is running.
 *
 * The loop test matters more than it looks. A bench that keeps ticking behind
 * a live view produces events with no device behind them, and they show up
 * later as a mysterious frame in the log that nothing sent.
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

function answerEverything() {
  mocked.gatewayStatus.mockResolvedValue(gatewayStatusFixture() as never);
  mocked.gatewayEvents.mockResolvedValue([] as never);
  mocked.gatewayHistory.mockResolvedValue({ clockSamples: [], healthOutcomes: [] } as never);
  mocked.gatewayFleet.mockResolvedValue({ summary: {}, devices: [] } as never);
  mocked.faultMode.mockResolvedValue(faultModeFixture() as never);
}

/** Mounts the engine without rendering anything of its own. */
function Engine() {
  useTelemetryEngine();
  return null;
}

beforeEach(() => {
  window.localStorage.clear();
  resetBenchStore();
  resetLiveStore();
  useSourceStore.setState({ source: "bench", hydrated: true });
  answerEverything();
});

describe("useTelemetry", () => {
  it("returns the bench by default", () => {
    const { result } = renderHook(() => useTelemetry());

    expect(result.current.source).toBe("bench");
    expect(result.current.editable).toBe(true);
  });

  it("returns the live source once selected", async () => {
    useSource("live");
    await act(async () => {
      await useLiveStore.getState().refresh();
    });

    const { result } = renderHook(() => useTelemetry());

    expect(result.current.source).toBe("live");
    expect(result.current.target).toBe("127.0.0.1:9000");
  });

  it("keeps a stable identity while nothing changes", () => {
    const { result, rerender } = renderHook(() => useTelemetry());
    const first = result.current;

    rerender();

    // A fresh object on every render makes zustand's snapshot comparison fail
    // and React reports it as an infinite loop.
    expect(result.current).toBe(first);
  });

  it("reports the active source on its own", () => {
    useSource("live");

    const { result } = renderHook(() => useSourceValue());

    expect(result.current).toBe("live");
  });
});

describe("useTelemetryEngine", () => {
  it("adopts the stored source on mount", () => {
    window.localStorage.setItem("ft12.source", "live");
    useSourceStore.setState({ source: "bench", hydrated: false });

    render(<Engine />);

    expect(useSourceStore.getState().source).toBe("live");
  });

  it("advances the bench while it is running", () => {
    jest.useFakeTimers();
    try {
      const tick = jest.spyOn(useBenchStore.getState(), "tick");
      useBenchStore.setState({ running: true });

      render(<Engine />);
      act(() => {
        jest.advanceTimersByTime(500);
      });

      expect(tick).toHaveBeenCalled();
    } finally {
      jest.useRealTimers();
    }
  });

  it("does not advance a stopped bench", () => {
    jest.useFakeTimers();
    try {
      const tick = jest.spyOn(useBenchStore.getState(), "tick");

      render(<Engine />);
      act(() => {
        jest.advanceTimersByTime(1000);
      });

      expect(tick).not.toHaveBeenCalled();
    } finally {
      jest.useRealTimers();
    }
  });

  it("leaves the bench alone while the live source is selected", () => {
    jest.useFakeTimers();
    try {
      useSource("live");
      useBenchStore.setState({ running: true });
      const tick = jest.spyOn(useBenchStore.getState(), "tick");

      render(<Engine />);
      act(() => {
        jest.advanceTimersByTime(1000);
      });

      // A bench ticking behind a live view produces frames nothing sent.
      expect(tick).not.toHaveBeenCalled();
    } finally {
      jest.useRealTimers();
    }
  });

  it("polls the API only while the live source is selected", async () => {
    render(<Engine />);
    expect(mocked.gatewayStatus).not.toHaveBeenCalled();

    act(() => useSource("live"));

    await waitFor(() => expect(mocked.gatewayStatus).toHaveBeenCalled());
    // Reads the emulator's faults once on entry, so the sliders start where
    // the device actually is rather than at zero.
    await waitFor(() => expect(mocked.faultMode).toHaveBeenCalled());
  });

  it("stops polling when unmounted", async () => {
    jest.useFakeTimers();
    try {
      useSource("live");
      const { unmount } = render(<Engine />);
      unmount();
      mocked.gatewayStatus.mockClear();

      act(() => {
        jest.advanceTimersByTime(5000);
      });

      expect(mocked.gatewayStatus).not.toHaveBeenCalled();
    } finally {
      jest.useRealTimers();
    }
  });
});

describe("useTelemetryControls", () => {
  it("drives the bench engine on the bench", () => {
    const { result } = renderHook(() => useTelemetryControls());

    act(() => result.current.start());

    expect(useBenchStore.getState().running).toBe(true);
    expect(result.current.reset).not.toBeNull();
  });

  it("commands the gateway on the live source", async () => {
    useSource("live");
    mocked.startPolling.mockResolvedValue(gatewayStatusFixture() as never);
    mocked.stopPolling.mockResolvedValue(gatewayStatusFixture() as never);

    const { result } = renderHook(() => useTelemetryControls());
    act(() => result.current.start());
    await waitFor(() => expect(mocked.startPolling).toHaveBeenCalled());

    act(() => result.current.stop());
    await waitFor(() => expect(mocked.stopPolling).toHaveBeenCalled());
  });

  it("offers no reset on the live source", () => {
    useSource("live");

    const { result } = renderHook(() => useTelemetryControls());

    // Clearing counters would mean asking a running gateway to forget what it
    // has measured, which is not something a toolbar should be able to do.
    expect(result.current.reset).toBeNull();
  });
});

describe("useFaultControls", () => {
  it("writes bench faults straight through", () => {
    const { result } = renderHook(() => useFaultControls());

    act(() => result.current.setFaults({ responseDelayMs: 700 }));

    expect(result.current.writable).toBe(true);
    expect(useBenchStore.getState().faults.responseDelayMs).toBe(700);
  });

  it("drops a null clock rather than writing it to the bench", () => {
    const { result } = renderHook(() => useFaultControls());

    act(() => result.current.setFaults({ noResponse: true, clockOffsetMs: null }));

    // Null only exists to say "the live source has no such control"; the
    // bench keeps whatever offset it had.
    expect(useBenchStore.getState().faults.clockOffsetMs).toBe(0);
    expect(useBenchStore.getState().faults.noResponse).toBe(true);
  });

  it("translates a probability into the emulator's flag and probability", async () => {
    useSource("live");
    await act(async () => {
      await useLiveStore.getState().loadFaults();
    });
    mocked.setFaultMode.mockResolvedValue(faultModeFixture() as never);

    const { result } = renderHook(() => useFaultControls());
    act(() => result.current.setFaults({ badChecksumProb: 0.35 }));

    await waitFor(() =>
      expect(mocked.setFaultMode).toHaveBeenCalledWith(
        expect.objectContaining({ corruptChecksumProbability: 0.35, corruptChecksum: false }),
      ),
    );
  });

  it("raises the emulator's flag for a fault on every frame", async () => {
    useSource("live");
    await act(async () => {
      await useLiveStore.getState().loadFaults();
    });
    mocked.setFaultMode.mockResolvedValue(faultModeFixture() as never);

    const { result } = renderHook(() => useFaultControls());
    act(() => result.current.setFaults({ fragmentProb: 1 }));

    await waitFor(() =>
      expect(mocked.setFaultMode).toHaveBeenCalledWith(
        expect.objectContaining({ fragmentProbability: 1, fragmentResponse: true }),
      ),
    );
  });

  it("sends only what changed", async () => {
    useSource("live");
    await act(async () => {
      await useLiveStore.getState().loadFaults();
    });
    mocked.setFaultMode.mockResolvedValue(faultModeFixture() as never);

    const { result } = renderHook(() => useFaultControls());
    act(() => result.current.setFaults({ noResponse: true }));

    await waitFor(() => expect(mocked.setFaultMode).toHaveBeenCalled());
    const sent = mocked.setFaultMode.mock.calls[0]![0];
    expect(sent.noResponse).toBe(true);
    // The clock fields have no counterpart on the emulator and are dropped.
    expect(sent).not.toHaveProperty("clockOffsetMs");
  });

  it("translates a whole preset in one write", async () => {
    useSource("live");
    await act(async () => {
      await useLiveStore.getState().loadFaults();
    });
    mocked.setFaultMode.mockResolvedValue(faultModeFixture() as never);

    const { result } = renderHook(() => useFaultControls());
    act(() =>
      result.current.setFaults({
        responseDelayMs: 900,
        badChecksumProb: 0,
        fragmentProb: 0,
        fragmentDelayMs: 60,
        noResponse: false,
        closeAfterRequest: true,
        clockOffsetMs: 0,
        clockDriftPerDayMs: 0,
      }),
    );

    await waitFor(() => expect(mocked.setFaultMode).toHaveBeenCalled());
    const sent = mocked.setFaultMode.mock.calls[0]![0];
    expect(sent).toMatchObject({
      responseDelayMs: 900,
      fragmentDelayMs: 60,
      noResponse: false,
      closeAfterRequest: true,
      corruptChecksumProbability: 0,
      corruptChecksum: false,
      fragmentProbability: 0,
      fragmentResponse: false,
    });
  });

  it("is not writable until the emulator has answered", () => {
    useSource("live");

    const { result } = renderHook(() => useFaultControls());

    // An operator dragging a slider that cannot reach anything deserves to be
    // told, not to watch it spring back.
    expect(result.current.writable).toBe(false);
  });
});
