"use client";

import { useMemo } from "react";

import { useFormat } from "@/lib/use-format";
import { cn } from "@/lib/utils";

/**
 * The bench's charts.
 *
 * All inline SVG, no charting library: every figure here is one series with a
 * known domain, and a library would add a bundle and a theming layer to draw
 * a polyline. It also means the axes can be exactly as sparse as an instrument
 * wants — a reading, its threshold, and nothing else.
 */

/** Sparkline for a counted series (frames per second, errors per second). */
export function Sparkline({
  values,
  className,
  tone = "primary",
  height = 40,
  label,
}: {
  values: number[];
  className?: string;
  tone?: "primary" | "warning" | "danger";
  height?: number;
  /** Announced in place of the drawing, which is silent on its own. */
  label?: string;
}) {
  const stroke = {
    primary: "var(--signal-rx)",
    warning: "var(--warning)",
    danger: "var(--destructive)",
  }[tone];

  const path = useMemo(() => {
    if (values.length < 2) return "";
    const max = Math.max(1, ...values);
    const step = 100 / (values.length - 1);
    return values
      .map((value, index) => {
        const x = index * step;
        const y = 100 - (value / max) * 100;
        return `${index === 0 ? "M" : "L"}${x.toFixed(2)},${y.toFixed(2)}`;
      })
      .join(" ");
  }, [values]);

  if (!path) {
    return <div className={cn("rounded bg-surface-sunken", className)} style={{ height }} />;
  }

  return (
    <svg
      viewBox="0 0 100 100"
      preserveAspectRatio="none"
      className={cn("w-full rounded bg-surface-sunken", className)}
      style={{ height }}
      role={label ? "img" : undefined}
      aria-label={label}
      aria-hidden={label ? undefined : true}
    >
      <path d={`${path} L100,100 L0,100 Z`} fill={stroke} opacity={0.12} />
      <path d={path} fill="none" stroke={stroke} strokeWidth={1.5} vectorEffect="non-scaling-stroke" />
    </svg>
  );
}

/**
 * Clock skew against its thresholds.
 *
 * A bare number cannot answer the question the operator actually has — "is
 * that a lot?" — so the scale carries the warn and critical bands it is being
 * judged against, and the needle sits where the reading falls between them.
 * The scale is symmetric because a device running fast and one running slow
 * are different faults, and losing the sign would hide which one this is.
 */
export function SkewMeter({
  skewMs,
  warnMs,
  criticalMs,
  samples,
}: {
  skewMs: number;
  warnMs: number;
  criticalMs: number;
  samples: number;
}) {
  const format = useFormat();

  // The scale runs to 1.4x critical so a device beyond the alarm still has
  // somewhere to sit rather than pinning silently at the edge.
  const limit = Math.max(criticalMs * 1.4, Math.abs(skewMs) * 1.1, 1000);
  const toPercent = (value: number) => 50 + (value / limit) * 50;

  const warnLeft = toPercent(-warnMs);
  const warnRight = toPercent(warnMs);
  const criticalLeft = toPercent(-criticalMs);
  const criticalRight = toPercent(criticalMs);
  const needle = Math.min(99.5, Math.max(0.5, toPercent(skewMs)));

  return (
    <div className="space-y-1.5">
      <div className="relative h-8 overflow-hidden rounded border border-border bg-surface-sunken">
        <div
          className="absolute inset-y-0 bg-destructive/18"
          style={{ left: 0, width: `${criticalLeft}%` }}
        />
        <div
          className="absolute inset-y-0 bg-destructive/18"
          style={{ left: `${criticalRight}%`, right: 0 }}
        />
        <div
          className="absolute inset-y-0 bg-warning/18"
          style={{ left: `${criticalLeft}%`, width: `${warnLeft - criticalLeft}%` }}
        />
        <div
          className="absolute inset-y-0 bg-warning/18"
          style={{ left: `${warnRight}%`, width: `${criticalRight - warnRight}%` }}
        />
        <div
          className="absolute inset-y-0 bg-success/14"
          style={{ left: `${warnLeft}%`, width: `${warnRight - warnLeft}%` }}
        />

        <div className="absolute inset-y-0 left-1/2 w-px bg-border-strong" />

        {samples > 0 ? (
          <div
            className="absolute inset-y-1 w-0.5 rounded-full bg-foreground shadow-[0_0_0_2px_var(--background)] transition-[left] duration-500 ease-out"
            style={{ left: `${needle}%` }}
          />
        ) : null}
      </div>
      <div className="flex justify-between font-mono text-[0.625rem] text-faint-foreground tabular">
        <span>-{format.duration(limit)}</span>
        <span>0</span>
        <span>+{format.duration(limit)}</span>
      </div>
    </div>
  );
}

/**
 * Skew history with the median drawn through it.
 *
 * Individual samples are dots rather than a line: the reading is noisy by
 * nature (line latency, one-second wire resolution) and joining the dots would
 * imply a precision the data does not have. The median line is the trend the
 * alarm actually acts on.
 */
export function SkewHistory({
  samples,
  medianMs,
  warnMs,
  height = 96,
}: {
  samples: { at: number; skewMs: number }[];
  medianMs: number;
  warnMs: number;
  height?: number;
}) {
  if (samples.length < 2) {
    return (
      <div
        className="flex items-center justify-center rounded bg-surface-sunken text-xs text-faint-foreground"
        style={{ height }}
      >
        —
      </div>
    );
  }

  const skews = samples.map((sample) => sample.skewMs);
  const bound = Math.max(Math.abs(Math.min(...skews)), Math.abs(Math.max(...skews)), warnMs) * 1.15;
  const step = 100 / (samples.length - 1);
  const toY = (value: number) => 50 - (value / bound) * 50;

  return (
    <div className="relative">
      <svg
        viewBox="0 0 100 100"
        preserveAspectRatio="none"
        className="w-full rounded bg-surface-sunken"
        style={{ height }}
        aria-hidden
      >
        <line x1="0" y1="50" x2="100" y2="50" stroke="var(--border-strong)" strokeWidth="0.4" />
        <line
          x1="0"
          y1={toY(warnMs)}
          x2="100"
          y2={toY(warnMs)}
          stroke="var(--warning)"
          strokeWidth="0.4"
          strokeDasharray="2 2"
          opacity={0.6}
        />
        <line
          x1="0"
          y1={toY(-warnMs)}
          x2="100"
          y2={toY(-warnMs)}
          stroke="var(--warning)"
          strokeWidth="0.4"
          strokeDasharray="2 2"
          opacity={0.6}
        />
        <line
          x1="0"
          y1={toY(medianMs)}
          x2="100"
          y2={toY(medianMs)}
          stroke="var(--signal-rx)"
          strokeWidth="0.8"
          vectorEffect="non-scaling-stroke"
        />
        {samples.map((sample, index) => (
          <circle
            key={sample.at + index}
            cx={index * step}
            cy={toY(sample.skewMs)}
            r={1.1}
            fill="var(--foreground)"
            opacity={0.55}
          />
        ))}
      </svg>
    </div>
  );
}

/**
 * Poll instants on a time axis.
 *
 * Aligned mode is a claim about *where* polls land, and the only honest way to
 * show a claim about phase is to draw the ticks and let the eye check that the
 * spacing is regular across a reconnect.
 */
export function ScheduleTimeline({
  ticks,
  windowMs,
  now,
}: {
  ticks: number[];
  windowMs: number;
  now: number;
}) {
  const format = useFormat();
  const start = now - windowMs;
  const visible = ticks.filter((tick) => tick >= start);

  return (
    <div className="relative h-10 overflow-hidden rounded border border-border bg-surface-sunken">
      {visible.map((tick, index) => {
        const left = ((tick - start) / windowMs) * 100;
        return (
          <div
            key={`${tick}-${index}`}
            className="absolute inset-y-2 w-px bg-signal-rx/70"
            style={{ left: `${left}%` }}
          />
        );
      })}
      <div className="absolute inset-y-0 right-0 w-px bg-primary" />
      <span className="absolute bottom-1 left-2 font-mono text-[0.625rem] text-faint-foreground tabular">
        -{format.duration(windowMs)}
      </span>
      <span className="absolute right-2 bottom-1 font-mono text-[0.625rem] text-faint-foreground tabular">
        now
      </span>
    </div>
  );
}

/** Availability as a run of outcome ticks — newest on the right. */
export function OutcomeStrip({ window: outcomes }: { window: boolean[] }) {
  const visible = outcomes.slice(-48);
  return (
    <div className="flex h-6 items-end gap-px">
      {visible.length === 0 ? (
        <div className="h-full w-full rounded bg-surface-sunken" />
      ) : (
        visible.map((ok, index) => (
          <div
            key={index}
            className={cn("min-w-0 flex-1 rounded-[1px]", ok ? "h-full bg-success/70" : "h-full bg-destructive/80")}
          />
        ))
      )}
    </div>
  );
}
