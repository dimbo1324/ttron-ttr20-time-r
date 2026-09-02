import { getDictionary } from "@/i18n";

import { createFormatter, SI_UNITS } from "./format";

const en = createFormatter("en", getDictionary("en").units);
const ru = createFormatter("ru", getDictionary("ru").units);

describe("duration in English", () => {
  it.each([
    { ms: 0.5, want: "0.50 ms" },
    { ms: 12, want: "12 ms" },
    { ms: 940, want: "940 ms" },
    { ms: 1420, want: "1.42 s" },
    { ms: 12_400, want: "12.4 s" },
    { ms: 125_000, want: "2 m 05 s" },
    { ms: 3_725_000, want: "1 h 02 m" },
  ])("renders $ms as $want", ({ ms, want }) => {
    expect(en.duration(ms)).toBe(want);
  });

  it("keeps the sign of a negative value", () => {
    expect(en.duration(-1420)).toBe("-1.42 s");
    expect(en.duration(-940)).toBe("-940 ms");
  });

  it("adds an explicit plus only when asked", () => {
    expect(en.duration(1420, { signed: true })).toBe("+1.42 s");
    expect(en.duration(1420)).toBe("1.42 s");
    expect(en.duration(-1420, { signed: true })).toBe("-1.42 s");
  });

  it("renders exactly zero without a sign or decimals", () => {
    expect(en.duration(0, { signed: true })).toBe("0 ms");
    expect(en.duration(0)).toBe("0 ms");
  });

  it("renders a non-finite value as a dash", () => {
    expect(en.duration(Number.NaN)).toBe("—");
    expect(en.duration(Number.POSITIVE_INFINITY)).toBe("—");
  });
});

describe("duration in Russian", () => {
  it("uses Russian unit symbols", () => {
    expect(ru.duration(940)).toBe("940 мс");
    expect(ru.duration(125_000)).toBe("2 мин 05 с");
    expect(ru.duration(3_725_000)).toBe("1 ч 02 мин");
  });

  it("uses a comma as the decimal separator", () => {
    expect(ru.duration(1420)).toContain(",");
    expect(ru.duration(1420)).toContain("с");
    expect(en.duration(1420)).toContain(".");
  });

  it("keeps the sign in both languages", () => {
    expect(ru.duration(-1420, { signed: true }).startsWith("-")).toBe(true);
    expect(ru.duration(1420, { signed: true }).startsWith("+")).toBe(true);
  });
});

describe("driftPerDay", () => {
  it("always carries its direction and its period", () => {
    expect(en.driftPerDay(24_000)).toBe("+24.0 s/day");
    expect(en.driftPerDay(-48_000)).toBe("-48.0 s/day");
    expect(en.driftPerDay(0)).toBe("0 s/day");
  });

  it.each([
    { ms: 500, want: "+500 ms/day" },
    { ms: 5000, want: "+5.00 s/day" },
    { ms: 120_000, want: "+2.0 m/day" },
  ])("renders $ms as $want", ({ ms, want }) => {
    expect(en.driftPerDay(ms)).toBe(want);
  });

  it("is localised", () => {
    expect(ru.driftPerDay(24_000)).toContain("с/сут");
    expect(ru.driftPerDay(500)).toContain("мс/сут");
  });

  it("renders a non-finite value as zero", () => {
    expect(en.driftPerDay(Number.NaN)).toBe("0 s/day");
  });
});

describe("percent", () => {
  it.each([
    { ratio: 1, digits: 1, want: "100.0%" },
    { ratio: 0, digits: 1, want: "0.0%" },
    { ratio: 0.5, digits: 0, want: "50%" },
    { ratio: 0.9876, digits: 2, want: "98.76%" },
  ])("renders $ratio as $want", ({ ratio, digits, want }) => {
    expect(en.percent(ratio, digits)).toBe(want);
  });

  it("renders a non-finite value as a dash", () => {
    expect(en.percent(Number.NaN)).toBe("—");
  });

  it("uses the Russian separator", () => {
    expect(ru.percent(0.5, 1)).toContain(",");
    expect(ru.percent(0.5, 1)).toContain("%");
  });
});

describe("clock", () => {
  const moment = new Date(2026, 8, 2, 17, 3, 7, 42).getTime();

  it("includes milliseconds by default", () => {
    expect(en.clock(moment)).toBe("17:03:07.042");
  });

  it("can omit milliseconds", () => {
    expect(en.clock(moment, false)).toBe("17:03:07");
  });

  it("pads every component", () => {
    expect(en.clock(new Date(2026, 0, 1, 1, 2, 3, 4).getTime())).toBe("01:02:03.004");
  });
});

describe("date", () => {
  it("is ISO in both locales, because 03-09 must not be a guess", () => {
    const at = new Date(2026, 8, 2).getTime();

    expect(en.date(at)).toBe("2026-09-02");
    expect(ru.date(at)).toBe("2026-09-02");
  });

  it("pads month and day", () => {
    expect(en.date(new Date(2026, 0, 5).getTime())).toBe("2026-01-05");
  });
});

describe("count", () => {
  it("groups thousands for the locale", () => {
    expect(en.count(1234567)).toBe("1,234,567");
    expect(ru.count(1234567)).not.toBe("1,234,567");
  });

  it("leaves small numbers alone", () => {
    expect(en.count(42)).toBe("42");
  });
});

describe("default units", () => {
  it("fall back to SI symbols when no dictionary is supplied", () => {
    const bare = createFormatter("en");

    expect(bare.duration(940)).toBe(`940 ${SI_UNITS.ms}`);
    expect(bare.driftPerDay(24_000)).toBe(`+24.0 ${SI_UNITS.s}${SI_UNITS.perDay}`);
  });
});
