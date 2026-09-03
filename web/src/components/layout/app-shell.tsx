"use client";

import {
  Activity,
  BookOpen,
  Binary,
  CircuitBoard,
  Gauge,
  Pause,
  Play,
  RotateCcw,
  Router,
  Waves,
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";

import { SourceSwitch } from "@/components/layout/source-switch";
import { useDictionary, useLocale } from "@/components/locale-provider";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { StateBadge } from "@/components/ui/state-badge";
import { LOCALE_SHORT, locales, withLocale } from "@/i18n";
import { useFormat } from "@/lib/use-format";
import { CLOCK_TONE } from "@/lib/status";
import {
  useTelemetry,
  useTelemetryControls,
  useTelemetryEngine,
} from "@/lib/telemetry/use-telemetry";
import { cn } from "@/lib/utils";

/**
 * The shell every page lives inside.
 *
 * Layout is a fixed rail plus a status strip, not a collapsing drawer: on a
 * bench the operator needs to reach any section in one click while a poll loop
 * is running, and a nav that has to be opened first costs a frame every time.
 *
 * The strip carries the two states that decide whether anything else on screen
 * can be trusted — is the device answering, and are its clocks sane — so they
 * are visible from every page rather than only from the dashboard.
 */

const NAV = [
  { href: "", key: "overview", icon: Gauge, section: "sectionWork" },
  { href: "/monitor", key: "monitor", icon: Waves, section: "sectionWork" },
  { href: "/protocol", key: "protocol", icon: Binary, section: "sectionWork" },
  { href: "/emulator", key: "emulator", icon: CircuitBoard, section: "sectionControl" },
  { href: "/gateway", key: "gateway", icon: Router, section: "sectionControl" },
  { href: "/reference", key: "reference", icon: BookOpen, section: "sectionLearn" },
] as const;

const SECTIONS = ["sectionWork", "sectionControl", "sectionLearn"] as const;

export function AppShell({ children }: { children: ReactNode }) {
  const dict = useDictionary();
  const locale = useLocale();
  const format = useFormat();
  const pathname = usePathname();

  // One mount drives whichever source is selected; see use-telemetry.
  useTelemetryEngine();
  const telemetry = useTelemetry();
  const { start, stop, reset, busy } = useTelemetryControls();

  const { running, clock } = telemetry;
  const health = telemetry.health.state;

  const base = `/${locale}`;

  return (
    <div className="flex min-h-screen bg-background bench-grid">
      <aside className="sticky top-0 flex h-screen w-[13.5rem] shrink-0 flex-col border-r border-border bg-surface/80 backdrop-blur">
        <Link href={base} className="flex items-center gap-2.5 border-b border-border px-4 py-3.5">
          <span className="flex size-7 items-center justify-center rounded border border-primary/40 bg-primary/10 text-primary">
            <Activity className="size-4" />
          </span>
          <span className="min-w-0">
            <span className="block truncate font-mono text-[0.8125rem] font-semibold text-foreground">
              {dict.shell.brand}
            </span>
            <span className="block truncate text-[0.6875rem] text-faint-foreground">
              {dict.shell.brandSub}
            </span>
          </span>
        </Link>

        <nav className="flex-1 overflow-y-auto px-2 py-3">
          {SECTIONS.map((section) => (
            <div key={section} className="mb-4 last:mb-0">
              <p className="eyebrow px-2 pb-1.5">{dict.nav[section]}</p>
              <ul className="space-y-0.5">
                {NAV.filter((item) => item.section === section).map((item) => {
                  const href = `${base}${item.href}`;
                  const active =
                    item.href === "" ? pathname === base || pathname === `${base}/` : pathname.startsWith(href);
                  const Icon = item.icon;
                  return (
                    <li key={item.key}>
                      <Link
                        href={href}
                        className={cn(
                          "flex items-center gap-2.5 rounded px-2 py-1.5 text-[0.8125rem] transition-colors duration-150",
                          active
                            ? "bg-primary/12 text-primary"
                            : "text-muted-foreground hover:bg-surface-raised hover:text-foreground",
                        )}
                      >
                        <Icon className="size-4 shrink-0" />
                        <span className="truncate">{dict.nav[item.key]}</span>
                      </Link>
                    </li>
                  );
                })}
              </ul>
            </div>
          ))}
        </nav>

        <div className="border-t border-border p-3">
          <p className="eyebrow pb-1.5">{dict.source.label}</p>
          <p className="text-[0.6875rem] leading-snug text-faint-foreground">
            {telemetry.source === "live" ? dict.source.liveHint : dict.source.benchHint}
          </p>
          <div className="mt-3 flex items-center gap-1">
            <span className="text-[0.6875rem] text-faint-foreground">{dict.shell.language}</span>
            <div className="ml-auto flex gap-0.5">
              {locales.map((item) => (
                <Link
                  key={item}
                  href={withLocale(pathname, item)}
                  className={cn(
                    "rounded px-1.5 py-0.5 font-mono text-[0.6875rem] transition-colors",
                    item === locale
                      ? "bg-surface-raised text-foreground"
                      : "text-faint-foreground hover:text-foreground",
                  )}
                >
                  {LOCALE_SHORT[item]}
                </Link>
              ))}
            </div>
          </div>
        </div>
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <header
          className="sticky top-0 z-20 flex h-14 items-center gap-3 border-b border-border bg-background/85 px-4 backdrop-blur"
          aria-label={dict.shell.statusStrip}
        >
          {/* The two states that decide whether anything else on screen can be
              trusted, announced when they change rather than only drawn. */}
          <div className="flex items-center gap-2" aria-live="polite">
            <StateBadge kind="health" state={health} pulse={running && health === "online"} />
            <Badge tone={CLOCK_TONE[clock.state]}>
              {dict.overview.skew}
              <span className="font-mono tabular">
                {clock.samples === 0
                  ? dict.common.none
                  : format.duration(clock.skewMs, { signed: true })}
              </span>
            </Badge>
          </div>

          <div className="ml-auto flex items-center gap-2">
            <SourceSwitch link={telemetry.link} />
            <Button
              size="sm"
              variant={running ? "outline" : "default"}
              disabled={busy || telemetry.link === "unreachable"}
              onClick={() => (running ? stop() : start())}
            >
              {running ? <Pause className="size-3.5" /> : <Play className="size-3.5" />}
              {running ? dict.common.stop : dict.common.start}
            </Button>
            {/* Reset clears the bench's own counters. The live gateway's are
                cleared only by restarting it, so the button is absent rather
                than present and inert. */}
            {reset ? (
              <Button size="iconSm" variant="ghost" onClick={reset} title={dict.common.reset}>
                <RotateCcw className="size-4" />
              </Button>
            ) : null}
          </div>
        </header>

        <main className="min-w-0 flex-1 p-4">{children}</main>
      </div>
    </div>
  );
}
