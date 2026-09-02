import {
  formatClock,
  formatCount,
  formatDate,
  formatDriftPerDay,
  formatDuration,
  formatPercent,
} from "./format";

describe("formatDuration", () => {
  it.each([
    { ms: 0.5, want: "0.50 ms" },
    { ms: 12, want: "12 ms" },
    { ms: 940, want: "940 ms" },
    { ms: 1420, want: "1.42 s" },
    { ms: 12_400, want: "12.4 s" },
    { ms: 125_000, want: "2m 05s" },
    { ms: 3_725_000, want: "1h 02m" },
  ])("renders $ms as $want", ({ ms, want }) => {
    expect(formatDuration(ms)).toBe(want);
  });

  it("keeps the sign of a negative value", () => {
    expect(formatDuration(-1420)).toBe("-1.42 s");
    expect(formatDuration(-940)).toBe("-940 ms");
  });

  it("adds an explicit plus only when asked", () => {
    expect(formatDuration(1420, { signed: true })).toBe("+1.42 s");
    expect(formatDuration(1420)).toBe("1.42 s");
    expect(formatDuration(-1420, { signed: true })).toBe("-1.42 s");
  });

  it("renders exactly zero without a sign or decimals", () => {
    expect(formatDuration(0, { signed: true })).toBe("0 ms");
    expect(formatDuration(0)).toBe("0 ms");
  });

  it("still uses two decimals below a millisecond", () => {
    expect(formatDuration(0.5)).toBe("0.50 ms");
  });

  it("renders a non-finite value as a dash", () => {
    expect(formatDuration(Number.NaN)).toBe("—");
    expect(formatDuration(Number.POSITIVE_INFINITY)).toBe("—");
  });
});

describe("formatDriftPerDay", () => {
  it.each([
    { ms: 0, want: "0 s" },
    { ms: 500, want: "+500 ms" },
    { ms: 24_000, want: "+24.0 s" },
    { ms: -48_000, want: "-48.0 s" },
    { ms: 120_000, want: "+2.0 m" },
  ])("renders $ms as $want", ({ ms, want }) => {
    expect(formatDriftPerDay(ms)).toBe(want);
  });

  it("renders a non-finite value as zero", () => {
    expect(formatDriftPerDay(Number.NaN)).toBe("0 s");
  });
});

describe("formatPercent", () => {
  it.each([
    { ratio: 1, digits: 1, want: "100.0%" },
    { ratio: 0, digits: 1, want: "0.0%" },
    { ratio: 0.5, digits: 0, want: "50%" },
    { ratio: 0.9876, digits: 2, want: "98.76%" },
  ])("renders $ratio as $want", ({ ratio, digits, want }) => {
    expect(formatPercent(ratio, digits)).toBe(want);
  });

  it("renders a non-finite value as a dash", () => {
    expect(formatPercent(Number.NaN)).toBe("—");
  });
});

describe("formatClock", () => {
  const moment = new Date(2026, 8, 2, 17, 3, 7, 42).getTime();

  it("includes milliseconds by default", () => {
    expect(formatClock(moment)).toBe("17:03:07.042");
  });

  it("can omit milliseconds", () => {
    expect(formatClock(moment, false)).toBe("17:03:07");
  });

  it("pads every component", () => {
    expect(formatClock(new Date(2026, 0, 1, 1, 2, 3, 4).getTime())).toBe("01:02:03.004");
  });
});

describe("formatDate", () => {
  it("renders an ISO calendar date", () => {
    expect(formatDate(new Date(2026, 8, 2).getTime())).toBe("2026-09-02");
  });

  it("pads month and day", () => {
    expect(formatDate(new Date(2026, 0, 5).getTime())).toBe("2026-01-05");
  });
});

describe("formatCount", () => {
  it("groups thousands", () => {
    expect(formatCount(1234567).replace(/ | /g, " ")).toBe("1 234 567");
  });

  it("leaves small numbers alone", () => {
    expect(formatCount(42)).toBe("42");
  });
});

describe("formatDriftPerDay precision", () => {
  it("uses two decimals below ten seconds", () => {
    expect(formatDriftPerDay(5000)).toBe("+5.00 s");
  });

  it("uses one decimal from ten seconds up", () => {
    expect(formatDriftPerDay(12_000)).toBe("+12.0 s");
  });
});
