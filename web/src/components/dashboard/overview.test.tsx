import { screen } from "@testing-library/react";

import { EMPTY_HEALTH, type HealthSnapshot } from "@/lib/bench/domain";
import { buildReadTimeRequest, encodeFrame } from "@/lib/ft12";
import { renderWithLocale, resetBenchStore } from "@/test/utils";

import { Overview } from "./overview";

const NOW = Date.UTC(2026, 8, 2, 12, 0, 0);

const healthy = (overrides: Partial<HealthSnapshot> = {}): HealthSnapshot => ({
  ...EMPTY_HEALTH,
  state: "online",
  successes: 12,
  window: new Array(12).fill(true),
  latencies: [8, 9, 10, 11, 12],
  consecutiveSuccesses: 12,
  ...overrides,
});

beforeEach(() => {
  jest.spyOn(Date, "now").mockReturnValue(NOW);
  resetBenchStore();
});

describe("Overview with no data", () => {
  it("shows both state machines as unknown", () => {
    const { dict } = renderWithLocale(<Overview />);

    expect(screen.getAllByText(dict.states.unknown).length).toBeGreaterThanOrEqual(2);
  });

  it("reports the poll loop as stopped", () => {
    const { dict } = renderWithLocale(<Overview />);

    expect(screen.getAllByText(dict.common.stopped).length).toBeGreaterThan(0);
  });

  it("names both services", () => {
    const { dict } = renderWithLocale(<Overview />);

    expect(screen.getByText(dict.overview.emulator)).toBeInTheDocument();
    expect(screen.getByText(dict.overview.gateway)).toBeInTheDocument();
  });

  it("calls a fault-free device healthy", () => {
    const { dict } = renderWithLocale(<Overview />);

    expect(screen.getByText(dict.emulator.presetHealthy)).toBeInTheDocument();
  });
});

describe("Overview with a healthy run", () => {
  beforeEach(() => {
    resetBenchStore({
      running: true,
      connected: true,
      health: healthy(),
      lastDeviceTime: NOW,
      lastCycleAt: NOW,
      nextPollAt: NOW + 3000,
      identityRead: { model: "TTR20", serial: "SN-0000042", firmware: "1.2.3" },
      samples: [0, 120, -80, 40].map((skewMs, index) => ({
        at: NOW - (4 - index) * 5000,
        skewMs,
        roundTripMs: 11,
      })),
      counters: {
        successfulReads: 12,
        failedReads: 0,
        reconnects: 0,
        retries: 0,
        protocolErrors: 0,
        exhaustedPolls: 0,
        connections: 1,
      },
      events: [
        {
          id: 1,
          at: NOW - 1000,
          direction: "rx",
          source: "device",
          command: "read-time",
          bytes: encodeFrame(0x80, 0x01, buildReadTimeRequest(), "sum"),
          cycle: 1,
          attempt: 0,
        },
      ],
    });
  });

  it("reports the clock as within tolerance", () => {
    const { dict } = renderWithLocale(<Overview />);

    expect(screen.getByText(dict.states.ok)).toBeInTheDocument();
  });

  it("reports full availability and the device online", () => {
    const { dict } = renderWithLocale(<Overview />);

    expect(screen.getByText("100.0%")).toBeInTheDocument();
    expect(screen.getByText(dict.states.online)).toBeInTheDocument();
  });

  it("shows latency percentiles", () => {
    renderWithLocale(<Overview />);

    expect(screen.getByText("p50")).toBeInTheDocument();
    expect(screen.getByText("p95")).toBeInTheDocument();
  });

  it("counts down to the next poll", () => {
    renderWithLocale(<Overview />);

    expect(screen.getByText("3.00 s")).toBeInTheDocument();
  });

  it("shows the device identity once it has been read", () => {
    renderWithLocale(<Overview />);

    expect(screen.getByText("TTR20")).toBeInTheDocument();
    expect(screen.getByText("SN-0000042")).toBeInTheDocument();
    expect(screen.getByText("1.2.3")).toBeInTheDocument();
  });

  it("names the aligned schedule", () => {
    const { dict } = renderWithLocale(<Overview />);

    expect(screen.getByText(dict.gateway.scheduleAligned)).toBeInTheDocument();
  });
});

describe("Overview with a drifting clock", () => {
  it("escalates to critical on a sustained skew", () => {
    resetBenchStore({
      running: true,
      samples: new Array(6).fill(null).map((_, index) => ({
        at: NOW - (6 - index) * 60_000,
        skewMs: 45_000,
        roundTripMs: 10,
      })),
    });
    const { dict } = renderWithLocale(<Overview />);

    expect(screen.getByText(dict.states.critical)).toBeInTheDocument();
  });

  it("reports a measured drift rate", () => {
    resetBenchStore({
      running: true,
      samples: new Array(5).fill(null).map((_, index) => ({
        at: NOW - (5 - index) * 3_600_000,
        skewMs: index * 1000,
        roundTripMs: 10,
      })),
    });
    renderWithLocale(<Overview />);

    expect(screen.getByText(/\/сутки/)).toBeInTheDocument();
    expect(screen.getByText(/R²/)).toBeInTheDocument();
  });
});

describe("Overview with an unhealthy device", () => {
  it("reports the device offline with zero availability", () => {
    resetBenchStore({
      running: true,
      health: {
        ...EMPTY_HEALTH,
        state: "offline",
        failures: 10,
        consecutiveFailures: 10,
        window: new Array(10).fill(false),
      },
      counters: {
        successfulReads: 0,
        failedReads: 10,
        reconnects: 2,
        retries: 4,
        protocolErrors: 3,
        exhaustedPolls: 10,
        connections: 3,
      },
    });
    const { dict } = renderWithLocale(<Overview />);

    expect(screen.getByText(dict.states.offline)).toBeInTheDocument();
    expect(screen.getByText("0.0%")).toBeInTheDocument();
  });

  it("counts the active faults on the emulator", () => {
    resetBenchStore({
      faults: {
        responseDelayMs: 500,
        badChecksumProb: 0.3,
        fragmentProb: 0,
        fragmentDelayMs: 40,
        noResponse: false,
        closeAfterRequest: false,
        clockOffsetMs: 4000,
        clockDriftPerDayMs: 0,
      },
    });
    const { dict } = renderWithLocale(<Overview />);

    expect(screen.getByText(`${dict.emulator.activeFaults}: 3`)).toBeInTheDocument();
  });
});
