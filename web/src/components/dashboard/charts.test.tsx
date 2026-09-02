import { screen } from "@testing-library/react";

import { renderWithLocale as render } from "@/test/utils";

import { OutcomeStrip, ScheduleTimeline, SkewHistory, SkewMeter, Sparkline } from "./charts";

const svgOf = (container: HTMLElement) => container.querySelector("svg");

describe("Sparkline", () => {
  it("draws a line and its fill for a series", () => {
    const { container } = render(<Sparkline values={[0, 2, 1, 4]} />);

    expect(svgOf(container)).toBeInTheDocument();
    expect(container.querySelectorAll("path")).toHaveLength(2);
  });

  it("falls back to a plain well for fewer than two points", () => {
    const { container } = render(<Sparkline values={[3]} />);

    expect(svgOf(container)).toBeNull();
    expect(container.firstChild).toHaveClass("bg-surface-sunken");
  });

  it("falls back for an empty series", () => {
    const { container } = render(<Sparkline values={[]} />);

    expect(svgOf(container)).toBeNull();
  });

  it("survives an all-zero series without dividing by zero", () => {
    const { container } = render(<Sparkline values={[0, 0, 0]} />);
    const path = container.querySelector("path")?.getAttribute("d") ?? "";

    expect(path).not.toContain("NaN");
  });

  it.each(["primary", "warning", "danger"] as const)("renders the %s tone", (tone) => {
    const { container } = render(<Sparkline values={[1, 2]} tone={tone} />);

    expect(svgOf(container)).toBeInTheDocument();
  });

  it("honours an explicit height", () => {
    const { container } = render(<Sparkline values={[1, 2]} height={80} />);

    expect(svgOf(container)).toHaveStyle({ height: "80px" });
  });
});

describe("SkewMeter", () => {
  it("labels a symmetric scale", () => {
    render(<SkewMeter skewMs={0} warnMs={2000} criticalMs={30_000} samples={3} />);

    expect(screen.getByText("0")).toBeInTheDocument();
    expect(screen.getByText(/^-/)).toBeInTheDocument();
    expect(screen.getByText(/^\+/)).toBeInTheDocument();
  });

  it("hides the needle until there is a sample", () => {
    const { container: empty } = render(
      <SkewMeter skewMs={0} warnMs={2000} criticalMs={30_000} samples={0} />,
    );
    const { container: measured } = render(
      <SkewMeter skewMs={1000} warnMs={2000} criticalMs={30_000} samples={1} />,
    );

    expect(empty.querySelectorAll("div[style*='left']").length).toBeLessThan(
      measured.querySelectorAll("div[style*='left']").length,
    );
  });

  it("keeps the needle inside the scale for an extreme reading", () => {
    const { container } = render(
      <SkewMeter skewMs={10_000_000} warnMs={2000} criticalMs={30_000} samples={5} />,
    );
    const needle = [...container.querySelectorAll<HTMLElement>("div[style*='left']")].at(-1);
    const left = Number.parseFloat(needle?.style.left ?? "0");

    expect(left).toBeGreaterThanOrEqual(0.5);
    expect(left).toBeLessThanOrEqual(99.5);
  });

  it("places a negative reading left of centre", () => {
    const { container } = render(
      <SkewMeter skewMs={-5000} warnMs={2000} criticalMs={30_000} samples={5} />,
    );
    const needle = [...container.querySelectorAll<HTMLElement>("div[style*='left']")].at(-1);

    expect(Number.parseFloat(needle?.style.left ?? "50")).toBeLessThan(50);
  });
});

describe("SkewHistory", () => {
  const samples = [0, 1000, 2000, 1500].map((skewMs, index) => ({
    at: index * 60_000,
    skewMs,
  }));

  it("draws one dot per sample plus the reference lines", () => {
    const { container } = render(<SkewHistory samples={samples} medianMs={1250} warnMs={2000} />);

    expect(container.querySelectorAll("circle")).toHaveLength(samples.length);
    expect(container.querySelectorAll("line")).toHaveLength(4);
  });

  it("shows a placeholder below two samples", () => {
    render(<SkewHistory samples={samples.slice(0, 1)} medianMs={0} warnMs={2000} />);

    expect(screen.getByText("—")).toBeInTheDocument();
  });

  it("keeps every dot inside the viewbox", () => {
    const { container } = render(<SkewHistory samples={samples} medianMs={1250} warnMs={2000} />);

    for (const circle of container.querySelectorAll("circle")) {
      const y = Number(circle.getAttribute("cy"));
      expect(y).toBeGreaterThanOrEqual(0);
      expect(y).toBeLessThanOrEqual(100);
    }
  });
});

describe("ScheduleTimeline", () => {
  const now = 1_000_000;

  it("draws a tick for every poll inside the window", () => {
    const { container } = render(
      <ScheduleTimeline ticks={[now - 50_000, now - 20_000, now - 5000]} windowMs={60_000} now={now} />,
    );

    // Three ticks plus the "now" edge.
    expect(container.querySelectorAll("div[class*='absolute']")).toHaveLength(4);
  });

  it("drops ticks older than the window", () => {
    const { container } = render(
      <ScheduleTimeline ticks={[now - 500_000, now - 5000]} windowMs={60_000} now={now} />,
    );

    expect(container.querySelectorAll("div[style*='left']")).toHaveLength(1);
  });

  it("renders an empty axis with no ticks", () => {
    render(<ScheduleTimeline ticks={[]} windowMs={60_000} now={now} />);

    expect(screen.getByText("now")).toBeInTheDocument();
  });
});

describe("OutcomeStrip", () => {
  it("renders one bar per outcome", () => {
    const { container } = render(<OutcomeStrip window={[true, false, true]} />);

    expect(container.firstChild?.childNodes).toHaveLength(3);
  });

  it("shows an empty well with no outcomes", () => {
    const { container } = render(<OutcomeStrip window={[]} />);

    expect(container.querySelector(".bg-surface-sunken")).toBeInTheDocument();
  });

  it("keeps only the most recent outcomes", () => {
    const { container } = render(<OutcomeStrip window={new Array(200).fill(true)} />);

    expect(container.firstChild?.childNodes.length).toBeLessThanOrEqual(48);
  });

  it("marks failures differently from successes", () => {
    const { container } = render(<OutcomeStrip window={[true, false]} />);
    const bars = [...(container.firstElementChild?.children ?? [])];

    expect(bars[0]?.className).toContain("bg-success");
    expect(bars[1]?.className).toContain("bg-destructive");
  });
});
