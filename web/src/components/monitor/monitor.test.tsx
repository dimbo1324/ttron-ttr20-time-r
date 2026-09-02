import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  buildReadIdentityResponse,
  buildReadTimeRequest,
  buildReadTimeResponse,
  encodeFrame,
  formatHex,
  RESPONSE_BIT,
} from "@/lib/ft12";
import { useBenchStore, type BenchEvent } from "@/stores/bench-store";
import { renderWithLocale, resetBenchStore } from "@/test/utils";

import { ExchangeMonitor } from "./monitor";

const NOW = Date.UTC(2026, 8, 2, 12, 0, 0);

const REQUEST = encodeFrame(0x00, 0x01, buildReadTimeRequest(), "sum");
const RESPONSE = encodeFrame(
  RESPONSE_BIT,
  0x01,
  buildReadTimeResponse(new Date(2026, 8, 2, 12, 0, 0)),
  "sum",
);
const IDENTITY = encodeFrame(
  RESPONSE_BIT,
  0x01,
  buildReadIdentityResponse("TTR20", "SN-42", "1.2.3"),
  "sum",
);

function event(overrides: Partial<BenchEvent> & Pick<BenchEvent, "id" | "direction">): BenchEvent {
  return {
    at: NOW,
    source: "gateway",
    command: "read-time",
    bytes: [],
    cycle: 1,
    attempt: 0,
    ...overrides,
  };
}

const LOG: BenchEvent[] = [
  event({ id: 1, direction: "tx", bytes: REQUEST }),
  event({ id: 2, direction: "rx", source: "device", bytes: RESPONSE, at: NOW + 12, latencyMs: 12 }),
  event({ id: 3, direction: "err", errorCode: "invalidChecksum", at: NOW + 20 }),
  event({
    id: 4,
    direction: "sys",
    command: "device-state",
    note: "deviceStateChanged",
    noteArgs: { from: "unknown", to: "online" },
    at: NOW + 30,
  }),
  event({ id: 5, direction: "rx", source: "device", command: "read-identity", bytes: IDENTITY, at: NOW + 40 }),
];

beforeEach(() => {
  jest.spyOn(Date, "now").mockReturnValue(NOW);
  resetBenchStore({ events: LOG });
});

describe("ExchangeMonitor log", () => {
  it("renders one row per event under a labelled header", () => {
    const { dict } = renderWithLocale(<ExchangeMonitor />);

    // The header is a row too — five unlabelled columns of hex are not a table.
    expect(screen.getAllByRole("row")).toHaveLength(LOG.length + 1);
    expect(screen.getByRole("columnheader", { name: dict.monitor.time })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: dict.monitor.raw })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: dict.monitor.latency })).toBeInTheDocument();
  });

  it("shows the raw frame of a protocol event", () => {
    renderWithLocale(<ExchangeMonitor />);

    expect(screen.getByText(formatHex(REQUEST))).toBeInTheDocument();
  });

  it("translates an error row", () => {
    const { dict } = renderWithLocale(<ExchangeMonitor />);

    expect(screen.getByText(dict.events.errors.invalidChecksum)).toBeInTheDocument();
  });

  it("renders a system event with both states translated", () => {
    const { dict } = renderWithLocale(<ExchangeMonitor />);

    expect(screen.getByText(dict.events.deviceStateChanged)).toBeInTheDocument();
    expect(
      screen.getByText(`${dict.states.unknown} → ${dict.states.online}`),
    ).toBeInTheDocument();
  });

  it("shows the measured latency of a response", () => {
    renderWithLocale(<ExchangeMonitor />);

    expect(screen.getByText("12 ms")).toBeInTheDocument();
  });

  it("prompts to start polling when the log is empty", () => {
    resetBenchStore({ events: [] });
    const { dict } = renderWithLocale(<ExchangeMonitor />);

    expect(screen.getByText(dict.monitor.empty)).toBeInTheDocument();
    expect(screen.getAllByText(dict.monitor.emptyHint).length).toBeGreaterThan(0);
  });
});

describe("ExchangeMonitor filters", () => {
  it("narrows the log to one direction", async () => {
    const { dict } = renderWithLocale(<ExchangeMonitor />);

    await userEvent.click(screen.getByRole("radio", { name: dict.directions.rx }));

    expect(screen.getAllByRole("row")).toHaveLength(3);
  });

  it("searches the hex and the command", async () => {
    const { dict } = renderWithLocale(<ExchangeMonitor />);
    const search = screen.getByPlaceholderText(dict.monitor.search);

    await userEvent.type(search, "read-identity");
    expect(screen.getAllByRole("row")).toHaveLength(2);

    await userEvent.clear(search);
    await userEvent.type(search, formatHex(REQUEST).slice(0, 8));
    expect(screen.getAllByRole("row").length).toBeGreaterThanOrEqual(1);
  });

  it("shows nothing when the search matches no frame", async () => {
    const { dict } = renderWithLocale(<ExchangeMonitor />);

    await userEvent.type(screen.getByPlaceholderText(dict.monitor.search), "zzzz");

    expect(screen.getByText(dict.monitor.empty)).toBeInTheDocument();
  });

  it("clears the log without touching the counters", async () => {
    const { dict } = renderWithLocale(<ExchangeMonitor />);

    await userEvent.click(screen.getByRole("button", { name: dict.monitor.clearLog }));

    expect(useBenchStore.getState().events).toEqual([]);
  });

  it("toggles auto-scroll", async () => {
    const { dict } = renderWithLocale(<ExchangeMonitor />);
    const toggle = screen.getByRole("button", { name: dict.monitor.autoscroll });

    await userEvent.click(toggle);

    expect(toggle).toBeInTheDocument();
  });
});

describe("ExchangeMonitor detail", () => {
  it("asks for a row to be selected", () => {
    const { dict } = renderWithLocale(<ExchangeMonitor />);

    expect(screen.getAllByText(dict.monitor.detailHint).length).toBeGreaterThan(0);
  });

  it("decodes the selected frame", async () => {
    const { dict } = renderWithLocale(<ExchangeMonitor />);

    await userEvent.click(screen.getByText(formatHex(RESPONSE)));

    expect(screen.getByText(dict.protocol.valid)).toBeInTheDocument();
    expect(screen.getByText("2026-09-02 12:00:00")).toBeInTheDocument();
  });

  it("explains a selected system event", async () => {
    const { dict } = renderWithLocale(<ExchangeMonitor />);

    await userEvent.click(screen.getByText(dict.events.deviceStateChanged));

    expect(screen.getAllByText("device-state").length).toBeGreaterThan(0);
    expect(screen.getAllByText(dict.events.deviceStateChanged).length).toBeGreaterThan(1);
  });
});

describe("ExchangeMonitor summary", () => {
  it("renders the sequence diagram for the frames it has", () => {
    const { container } = renderWithLocale(<ExchangeMonitor />);

    expect(container.querySelectorAll("div[style*='left']").length).toBeGreaterThan(0);
  });

  it("shows the counters", () => {
    resetBenchStore({
      events: LOG,
      counters: {
        successfulReads: 5,
        failedReads: 1,
        reconnects: 2,
        retries: 3,
        protocolErrors: 4,
        exhaustedPolls: 1,
        connections: 6,
      },
    });
    const { dict } = renderWithLocale(<ExchangeMonitor />);
    const counters = screen.getByRole("heading", { name: dict.monitor.counters }).closest("section");

    expect(within(counters!).getByText("5")).toBeInTheDocument();
    expect(within(counters!).getByText("4")).toBeInTheDocument();
  });
});
