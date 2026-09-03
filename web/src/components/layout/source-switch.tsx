"use client";

import { CircuitBoard, Radio } from "lucide-react";

import { useDictionary } from "@/components/locale-provider";
import { Badge } from "@/components/ui/badge";
import { StatusDot } from "@/components/ui/status-dot";
import type { LinkState, TelemetrySource } from "@/lib/telemetry/types";
import { cn } from "@/lib/utils";
import { useSourceStore } from "@/stores/source-store";

/**
 * Which stack the console is reading.
 *
 * This is the one control on the page that changes what every number means, so
 * it is a two-position switch with both options permanently visible rather
 * than a menu: an operator has to be able to tell, without clicking anything,
 * whether the reading in front of them came from a simulation or from a meter.
 *
 * The link indicator sits inside the switch for the same reason. "Live" and
 * "live but nothing is answering" look identical in a status bar somewhere
 * else on the page, and confusing them is how a dead bench gets read as a
 * healthy one.
 */

const OPTIONS = [
  { value: "bench", icon: CircuitBoard },
  { value: "live", icon: Radio },
] as const satisfies readonly { value: TelemetrySource; icon: typeof Radio }[];

const LINK_TONE = {
  ready: "success",
  connecting: "neutral",
  unreachable: "danger",
} as const satisfies Record<LinkState, "success" | "neutral" | "danger">;

const LINK_DOT = {
  ready: "online",
  connecting: "idle",
  unreachable: "offline",
} as const satisfies Record<LinkState, "online" | "idle" | "offline">;

export function SourceSwitch({ link }: { link: LinkState }) {
  const dict = useDictionary();
  const source = useSourceStore((state) => state.source);
  const setSource = useSourceStore((state) => state.setSource);

  return (
    <div className="flex items-center gap-2">
      <div
        role="radiogroup"
        aria-label={dict.source.label}
        className="flex rounded border border-border bg-surface-sunken p-0.5"
      >
        {OPTIONS.map((option) => {
          const active = source === option.value;
          const Icon = option.icon;
          return (
            <button
              key={option.value}
              type="button"
              role="radio"
              aria-checked={active}
              // The hint is a tooltip, not the name: without an explicit label
              // the title attribute becomes the accessible name, and a screen
              // reader announces a paragraph where it should say "Bench".
              aria-label={dict.source[option.value]}
              title={dict.source[optionHintKey(option.value)]}
              onClick={() => setSource(option.value)}
              className={cn(
                "flex items-center gap-1.5 rounded-[3px] px-2 py-1 text-[0.6875rem] font-medium transition-colors duration-150",
                active
                  ? "bg-primary/15 text-primary"
                  : "text-faint-foreground hover:text-foreground",
              )}
            >
              <Icon className="size-3.5" />
              {dict.source[option.value]}
            </button>
          );
        })}
      </div>

      {source === "live" ? (
        <Badge tone={LINK_TONE[link]}>
          <StatusDot tone={LINK_DOT[link]} pulse={link === "connecting"} />
          {dict.source[linkKey(link)]}
        </Badge>
      ) : null}
    </div>
  );
}

/** Keys a source onto its hint without a cast at the call site. */
function optionHintKey(source: TelemetrySource): "benchHint" | "liveHint" {
  return source === "live" ? "liveHint" : "benchHint";
}

/** Keys the link states into the dictionary without a cast at the call site. */
function linkKey(link: LinkState): "linkReady" | "linkConnecting" | "linkUnreachable" {
  if (link === "ready") return "linkReady";
  if (link === "connecting") return "linkConnecting";
  return "linkUnreachable";
}
