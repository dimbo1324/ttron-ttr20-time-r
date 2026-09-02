import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import type { ClockState, HealthState } from "@/lib/bench/domain";
import { CLOCK_DOT, CLOCK_TONE, DIRECTION_BG, DIRECTION_TEXT, DIRECTION_TONE, HEALTH_DOT, HEALTH_TONE } from "@/lib/status";
import { useNow } from "@/lib/use-now";
import { renderWithLocale } from "@/test/utils";

import { CopyButton } from "./copy-button";
import { StateBadge } from "./state-badge";

const CLOCK_STATES: ClockState[] = ["unknown", "ok", "warn", "critical"];
const HEALTH_STATES: HealthState[] = ["unknown", "online", "degraded", "offline"];

describe("status maps", () => {
  it("colour every clock state", () => {
    for (const state of CLOCK_STATES) {
      expect(CLOCK_TONE[state]).toBeTruthy();
      expect(CLOCK_DOT[state]).toBeTruthy();
    }
  });

  it("colour every health state", () => {
    for (const state of HEALTH_STATES) {
      expect(HEALTH_TONE[state]).toBeTruthy();
      expect(HEALTH_DOT[state]).toBeTruthy();
    }
  });

  it("keep the four directions distinct", () => {
    const directions = ["tx", "rx", "err", "sys"] as const;

    expect(new Set(directions.map((d) => DIRECTION_TONE[d])).size).toBe(4);
    expect(new Set(directions.map((d) => DIRECTION_TEXT[d])).size).toBe(4);
    expect(new Set(directions.map((d) => DIRECTION_BG[d])).size).toBe(4);
  });

  it("agree on what a healthy reading looks like", () => {
    // The two machines share a vocabulary of colour: "ok" and "online" are one
    // idea to the reader, and the maps must not disagree about it.
    expect(CLOCK_TONE.ok).toBe(HEALTH_TONE.online);
    expect(CLOCK_TONE.critical).toBe(HEALTH_TONE.offline);
    expect(CLOCK_TONE.unknown).toBe(HEALTH_TONE.unknown);
  });
});

describe("StateBadge", () => {
  it.each(CLOCK_STATES)("names the %s clock state", (state) => {
    const { dict } = renderWithLocale(<StateBadge kind="clock" state={state} />);

    expect(screen.getByText(dict.states[state])).toBeInTheDocument();
  });

  it.each(HEALTH_STATES)("names the %s device state", (state) => {
    const { dict } = renderWithLocale(<StateBadge kind="health" state={state} />);

    expect(screen.getByText(dict.states[state])).toBeInTheDocument();
  });

  it("pulses only when told to", () => {
    const { container: still } = renderWithLocale(<StateBadge kind="health" state="online" />);
    const { container: live } = renderWithLocale(
      <StateBadge kind="health" state="online" pulse />,
    );

    expect(still.querySelector(".pulse-dot")).toBeNull();
    expect(live.querySelector(".pulse-dot")).not.toBeNull();
  });

  it("translates with the locale", () => {
    const { dict } = renderWithLocale(<StateBadge kind="clock" state="warn" />, { locale: "en" });

    expect(screen.getByText(dict.states.warn)).toBeInTheDocument();
  });
});

describe("CopyButton", () => {
  it("copies its value and confirms", async () => {
    const writeText = jest.fn(async () => undefined);
    Object.assign(navigator.clipboard, { writeText });

    const { dict } = renderWithLocale(<CopyButton value="68 03 68" />);

    await userEvent.click(screen.getByRole("button", { name: dict.common.copy }));

    expect(writeText).toHaveBeenCalledWith("68 03 68");
    await waitFor(() =>
      expect(screen.getByRole("button", { name: dict.common.copied })).toBeInTheDocument(),
    );
  });

  it("returns to its resting label", async () => {
    // Real timers: the confirmation is set from a resolved promise and cleared
    // from a timeout, and driving both through fake timers makes the test more
    // intricate than the component it is checking.
    Object.assign(navigator.clipboard, { writeText: jest.fn(async () => undefined) });
    const { dict } = renderWithLocale(<CopyButton value="68" />);

    await userEvent.click(screen.getByRole("button"));
    await waitFor(() =>
      expect(screen.getByRole("button", { name: dict.common.copied })).toBeInTheDocument(),
    );

    await waitFor(
      () => expect(screen.getByRole("button", { name: dict.common.copy })).toBeInTheDocument(),
      { timeout: 2500 },
    );
  });

  it("stays quiet when the clipboard refuses", async () => {
    Object.assign(navigator.clipboard, {
      writeText: jest.fn(async () => {
        throw new Error("denied");
      }),
    });

    const { dict } = renderWithLocale(<CopyButton value="68" />);

    await userEvent.click(screen.getByRole("button"));

    // The API is unavailable over plain HTTP on a remote origin; a refusal is
    // expected rather than exceptional, so nothing throws and nothing confirms.
    expect(screen.getByRole("button", { name: dict.common.copy })).toBeInTheDocument();
  });

  it("uses a caller-supplied label", () => {
    renderWithLocale(<CopyButton value="68" label="Скопировать кадр" />);

    expect(screen.getByRole("button", { name: "Скопировать кадр" })).toBeInTheDocument();
  });

  it("is disabled with nothing to copy", () => {
    renderWithLocale(<CopyButton value="" />);

    expect(screen.getByRole("button")).toBeDisabled();
  });
});

describe("useNow", () => {
  function Probe() {
    return <span data-testid="now">{useNow(50)}</span>;
  }

  it("starts at zero so the server and the first client render agree", () => {
    let first: number | undefined;

    function Capture() {
      const now = useNow();
      first ??= now;
      return null;
    }

    render(<Capture />);
    expect(first).toBe(0);
  });

  it("advances after mount", async () => {
    render(<Probe />);

    await waitFor(() => {
      expect(Number(screen.getByTestId("now").textContent)).toBeGreaterThan(0);
    });
  });

  it("stops ticking once unmounted", () => {
    jest.useFakeTimers();
    try {
      const clearInterval = jest.spyOn(window, "clearInterval");
      const { unmount } = render(<Probe />);

      unmount();

      expect(clearInterval).toHaveBeenCalled();
    } finally {
      jest.useRealTimers();
    }
  });
});
