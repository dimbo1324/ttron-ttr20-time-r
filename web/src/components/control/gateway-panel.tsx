"use client";

import { CalendarClock, Repeat, Router, ShieldAlert } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import { ScheduleTimeline, SkewHistory } from "@/components/dashboard/charts";
import { useDictionary } from "@/components/locale-provider";
import { Badge } from "@/components/ui/badge";
import { Field, SegmentedControl, Select } from "@/components/ui/controls";
import { DefRow, Panel, PanelBody, PanelHeader, Stat } from "@/components/ui/panel";
import { StatusDot } from "@/components/ui/status-dot";
import type { ScheduleMode } from "@/lib/bench/domain";
import { formatDuration } from "@/lib/format";
import { useMounted } from "@/lib/use-mounted";
import { useBenchStore, useClockReport } from "@/stores/bench-store";

/**
 * Gateway control.
 *
 * The schedule section is the one that needs showing rather than describing:
 * the difference between "every 5 seconds" and "on the fifth second of every
 * minute" is invisible in a form and obvious on a timeline, so the ticks sit
 * directly under the control that produces them.
 */

const INTERVALS = [1000, 2000, 5000, 10_000, 30_000, 60_000];
const OFFSETS = [0, 1000, 2000, 5000, 10_000, 15_000, 30_000];
const TIMEOUTS = [250, 500, 1000, 1500, 3000];
const RETRIES = [0, 1, 2, 3, 5];
const WARN_THRESHOLDS = [500, 1000, 2000, 5000, 10_000];
const CRITICAL_THRESHOLDS = [5000, 10_000, 30_000, 60_000, 300_000];

export function GatewayPanel() {
  const dict = useDictionary();

  const running = useBenchStore((state) => state.running);
  const connected = useBenchStore((state) => state.connected);
  const gateway = useBenchStore((state) => state.gateway);
  const counters = useBenchStore((state) => state.counters);
  const health = useBenchStore((state) => state.health);
  const samples = useBenchStore((state) => state.samples);
  const events = useBenchStore((state) => state.events);
  const nextPoll = useBenchStore((state) => state.nextPollAt);
  const patchGateway = useBenchStore((state) => state.patchGateway);
  const clock = useClockReport();

  const mounted = useMounted();
  const [now, setNow] = useState(0);
  useEffect(() => {
    setNow(Date.now());
    const id = window.setInterval(() => setNow(Date.now()), 250);
    return () => window.clearInterval(id);
  }, []);

  // One tick per request that opened a cycle — retries inside a cycle are not
  // schedule events and would make a regular schedule look ragged.
  const ticks = useMemo(
    () =>
      events
        .filter((event) => event.direction === "tx" && event.attempt === 0 && event.command === "read-time")
        .map((event) => event.at),
    [events],
  );

  const alignedHint =
    gateway.scheduleMode === "aligned"
      ? dict.gateway.scheduleAlignedHint
      : dict.gateway.scheduleIntervalHint;

  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_23rem]">
      <div className="space-y-4">
        <Panel>
          <PanelHeader
            icon={<CalendarClock className="size-4" />}
            title={dict.gateway.title}
            hint={dict.gateway.subtitle}
            actions={
              <Badge tone={running ? "success" : "neutral"}>
                <StatusDot tone={running ? "online" : "idle"} pulse={running} />
                {running ? dict.gateway.pollingRunning : dict.gateway.pollingStopped}
              </Badge>
            }
          />
          <PanelBody className="space-y-3.5">
            <Field label={dict.gateway.scheduleMode} hint={alignedHint}>
              <SegmentedControl<ScheduleMode>
                value={gateway.scheduleMode}
                onChange={(mode) => patchGateway({ scheduleMode: mode })}
                options={[
                  { value: "interval", label: dict.gateway.scheduleInterval },
                  { value: "aligned", label: dict.gateway.scheduleAligned },
                ]}
                className="w-full"
              />
            </Field>

            <div className="grid gap-3 sm:grid-cols-2">
              <Field label={dict.gateway.interval}>
                <Select
                  value={String(gateway.intervalMs)}
                  onChange={(event) => patchGateway({ intervalMs: Number(event.target.value) })}
                >
                  {INTERVALS.map((value) => (
                    <option key={value} value={value}>
                      {formatDuration(value)}
                    </option>
                  ))}
                </Select>
              </Field>
              <Field
                label={dict.gateway.offset}
                hint={gateway.scheduleMode === "interval" ? dict.gateway.scheduleIntervalHint : undefined}
              >
                <Select
                  value={String(gateway.offsetMs)}
                  disabled={gateway.scheduleMode !== "aligned"}
                  onChange={(event) => patchGateway({ offsetMs: Number(event.target.value) })}
                >
                  {OFFSETS.filter((value) => value < gateway.intervalMs).map((value) => (
                    <option key={value} value={value}>
                      {value === 0 ? "0" : formatDuration(value)}
                    </option>
                  ))}
                </Select>
              </Field>
            </div>

            <div>
              <p className="eyebrow pb-1.5">{dict.gateway.timeline}</p>
              <ScheduleTimeline ticks={ticks} windowMs={60_000} now={now} />
              <p className="mt-1.5 text-xs text-faint-foreground">{dict.gateway.timelineHint}</p>
            </div>
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader
            icon={<Repeat className="size-4" />}
            title={dict.gateway.retryAttempts}
            hint={dict.gateway.retryHint}
          />
          <PanelBody className="grid gap-3 sm:grid-cols-2">
            <Field label={dict.gateway.requestTimeout}>
              <Select
                value={String(gateway.requestTimeoutMs)}
                onChange={(event) => patchGateway({ requestTimeoutMs: Number(event.target.value) })}
              >
                {TIMEOUTS.map((value) => (
                  <option key={value} value={value}>
                    {formatDuration(value)}
                  </option>
                ))}
              </Select>
            </Field>
            <Field label={dict.gateway.retryAttempts}>
              <Select
                value={String(gateway.retryAttempts)}
                onChange={(event) => patchGateway({ retryAttempts: Number(event.target.value) })}
              >
                {RETRIES.map((value) => (
                  <option key={value} value={value}>
                    {value}
                  </option>
                ))}
              </Select>
            </Field>
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader
            icon={<ShieldAlert className="size-4" />}
            title={dict.gateway.clockThresholds}
            hint={dict.overview.skewHint}
          />
          <PanelBody className="space-y-3">
            <div className="grid gap-3 sm:grid-cols-2">
              <Field label={dict.gateway.warnThreshold}>
                <Select
                  value={String(gateway.thresholds.warnMs)}
                  onChange={(event) =>
                    patchGateway({
                      thresholds: { ...gateway.thresholds, warnMs: Number(event.target.value) },
                    })
                  }
                >
                  {WARN_THRESHOLDS.map((value) => (
                    <option key={value} value={value}>
                      {formatDuration(value)}
                    </option>
                  ))}
                </Select>
              </Field>
              <Field label={dict.gateway.criticalThreshold}>
                <Select
                  value={String(gateway.thresholds.criticalMs)}
                  onChange={(event) =>
                    patchGateway({
                      thresholds: { ...gateway.thresholds, criticalMs: Number(event.target.value) },
                    })
                  }
                >
                  {CRITICAL_THRESHOLDS.map((value) => (
                    <option key={value} value={value}>
                      {formatDuration(value)}
                    </option>
                  ))}
                </Select>
              </Field>
            </div>
            <div>
              <p className="eyebrow pb-1.5">{dict.gateway.skewChart}</p>
              <SkewHistory
                samples={samples}
                medianMs={clock.medianMs}
                warnMs={gateway.thresholds.warnMs}
                height={110}
              />
              <p className="mt-1.5 text-xs text-faint-foreground">{dict.gateway.skewChartHint}</p>
            </div>
          </PanelBody>
        </Panel>
      </div>

      <div className="space-y-4">
        <Panel className="xl:sticky xl:top-[4.5rem]">
          <PanelHeader icon={<Router className="size-4" />} title={dict.gateway.connection} />
          <PanelBody className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <Stat
                label={dict.overview.nextPoll}
                value={mounted && running ? formatDuration(Math.max(0, nextPoll - now)) : dict.common.stopped}
                mono
                tone={running ? "primary" : "muted"}
              />
              <Stat
                label={dict.gateway.connection}
                value={connected ? dict.common.connected : dict.common.disconnected}
                tone={connected ? "success" : "muted"}
              />
            </div>
            <div className="border-t border-border pt-2">
              <DefRow label={dict.overview.successfulReads} value={counters.successfulReads} />
              <DefRow label={dict.overview.failedReads} value={counters.failedReads} />
              <DefRow label={dict.overview.retries} value={counters.retries} />
              <DefRow label={dict.overview.protocolErrors} value={counters.protocolErrors} />
              <DefRow label={dict.overview.reconnects} value={counters.reconnects} />
              <DefRow label={dict.overview.sessions} value={counters.connections} />
            </div>
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader title={dict.gateway.healthPolicy} />
          <PanelBody className="space-y-3">
            <div className="grid grid-cols-3 gap-2">
              <Field label={dict.gateway.degradeAfter}>
                <Select
                  value={String(gateway.policy.degradeAfter)}
                  onChange={(event) =>
                    patchGateway({
                      policy: { ...gateway.policy, degradeAfter: Number(event.target.value) },
                    })
                  }
                >
                  {[1, 2, 3, 5, 8].map((value) => (
                    <option key={value} value={value}>
                      {value}
                    </option>
                  ))}
                </Select>
              </Field>
              <Field label={dict.gateway.offlineAfter}>
                <Select
                  value={String(gateway.policy.offlineAfter)}
                  onChange={(event) =>
                    patchGateway({
                      policy: { ...gateway.policy, offlineAfter: Number(event.target.value) },
                    })
                  }
                >
                  {[3, 5, 10, 15, 20].map((value) => (
                    <option key={value} value={value}>
                      {value}
                    </option>
                  ))}
                </Select>
              </Field>
              <Field label={dict.gateway.recoverAfter}>
                <Select
                  value={String(gateway.policy.recoverAfter)}
                  onChange={(event) =>
                    patchGateway({
                      policy: { ...gateway.policy, recoverAfter: Number(event.target.value) },
                    })
                  }
                >
                  {[1, 2, 3, 5].map((value) => (
                    <option key={value} value={value}>
                      {value}
                    </option>
                  ))}
                </Select>
              </Field>
            </div>
            <div className="border-t border-border pt-2">
              <DefRow
                label={dict.gateway.failuresShort}
                value={health.consecutiveFailures}
              />
              <DefRow
                label={dict.gateway.successesShort}
                value={health.consecutiveSuccesses}
              />
              <DefRow
                label={dict.states.online}
                value={dict.states[health.state as keyof typeof dict.states] ?? dict.states.unknown}
              />
            </div>
          </PanelBody>
        </Panel>
      </div>
    </div>
  );
}
