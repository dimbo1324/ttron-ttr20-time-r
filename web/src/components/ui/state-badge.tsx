"use client";

import { useDictionary } from "@/components/locale-provider";
import { Badge } from "@/components/ui/badge";
import { StatusDot } from "@/components/ui/status-dot";
import type { ClockState, HealthState } from "@/lib/bench/domain";
import { CLOCK_DOT, CLOCK_TONE, HEALTH_DOT, HEALTH_TONE } from "@/lib/status";

/**
 * A domain state, rendered the same way everywhere it appears.
 *
 * Dot, tone and translated label always travel together — a coloured badge
 * with no word is unreadable to anyone who has not learnt the palette, and a
 * word with no colour is invisible at a glance. Binding the three here is what
 * stops one of them being forgotten at a call site.
 */
export function StateBadge(
  props: ({ kind: "clock"; state: ClockState } | { kind: "health"; state: HealthState }) & {
    pulse?: boolean;
  },
) {
  const dict = useDictionary();
  const { pulse = false } = props;

  const tone = props.kind === "clock" ? CLOCK_TONE[props.state] : HEALTH_TONE[props.state];
  const dot = props.kind === "clock" ? CLOCK_DOT[props.state] : HEALTH_DOT[props.state];

  return (
    <Badge tone={tone}>
      <StatusDot tone={dot} pulse={pulse} />
      {dict.states[props.state]}
    </Badge>
  );
}
