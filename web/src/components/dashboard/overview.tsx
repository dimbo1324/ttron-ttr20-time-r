"use client";

import { Clock, CircuitBoard, Gauge, Router, Waves } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import { useDictionary } from "@/components/locale-provider";
import { OutcomeStrip, SkewHistory, SkewMeter, Sparkline } from "@/components/dashboard/charts";
import { Badge } from "@/components/ui/badge";
import { Panel, PanelBody, PanelHeader, DefRow, Stat } from "@/components/ui/panel";
import { StatusDot, type StatusTone } from "@/components/ui/status-dot";
import { availability, percentile } from "@/lib/bench/domain";
import { useMounted } from "@/lib/use-mounted";
import { formatClock, formatDriftPerDay, formatDuration, formatPercent } from "@/lib/format";
import { formatDeviceTime } from "@/lib/ft12";
import { selectActiveFaultCount, useBenchStore, useClockReport } from "@/stores/bench-store";

/**
 * The overview.
 *
 * Ordered by the question an operator asks first and the one they ask last:
 * is the device answering, is its clock trustworthy, and only then how the
 * line and the counters look. Everything above the fold is a state; the
 * numbers that explain a state sit under it.
 */
export function Overview() {
  const dict = useDictionary();

  const running = useBenchStore((state) => state.running);
  const connected = useBenchStore((state) => state.connected);
  const events = useBenchStore((state) => state.events);
  const samples = useBenchStore((state) => state.samples);
  const health = useBenchStore((state) => state.health);
  const counters = useBenchStore((state) => state.counters);
  const gateway = useBenchStore((state) => state.gateway);
  const faults = useBenchStore((state) => state.faults);
  const checksumMode = useBenchStore((state) => state.checksumMode);
  const lastDeviceTime = useBenchStore((state) => state.lastDeviceTime);
  const identityRead = useBenchStore((state) => state.identityRead);
  const nextPoll = useBenchStore((state) => state.nextPollAt);
  const lastCycleAt = useBenchStore((state) => state.lastCycleAt);
  const activeFaults = useBenchStore(selectActiveFaultCount);
  const clock = useClockReport();

  // A ticking readout needs its own clock: the store only changes when a poll
  // resolves, and a countdown that only moves every five seconds is not one.
  // It starts at 0 so the server and the first client render agree — the real
  // time only arrives after mount.
  const mounted = useMounted();
  const [now, setNow] = useState(0);
  useEffect(() => {
    setNow(Date.now());
    const id = window.setInterval(() => setNow(Date.now()), 250);
    return () => window.clearInterval(id);
  }, []);

  const frameRate = useMemo(() => {
    const buckets = new Array(60).fill(0) as number[];
    const start = now - 60_000;
    for (const event of events) {
      if (event.at < start || event.bytes.length === 0) continue;
      const index = Math.min(59, Math.floor((event.at - start) / 1000));
      buckets[index] = (buckets[index] ?? 0) + 1;
    }
    return buckets;
  }, [events, now]);

  const clockTone: StatusTone =
    clock.state === "ok"
      ? "online"
      : clock.state === "warn"
        ? "degraded"
        : clock.state === "critical"
          ? "offline"
          : "unknown";

  const healthTone: StatusTone =
    health.state === "online"
      ? "online"
      : health.state === "degraded"
        ? "degraded"
        : health.state === "offline"
          ? "offline"
          : "unknown";

  const countdown = Math.max(0, nextPoll - now);
  const latencies = health.latencies;

  return (
    <div className="space-y-4">
      <div className="grid gap-4 lg:grid-cols-3">
        <Panel>
          <PanelHeader
            icon={<Clock className="size-4" />}
            title={dict.overview.skew}
            hint={dict.overview.skewHint}
            actions={
              <Badge
                tone={
                  clock.state === "ok"
                    ? "success"
                    : clock.state === "warn"
                      ? "warning"
                      : clock.state === "critical"
                        ? "danger"
                        : "neutral"
                }
              >
                <StatusDot tone={clockTone} />
                {dict.states[clock.state as keyof typeof dict.states] ?? dict.states.unknown}
              </Badge>
            }
          />
          <PanelBody className="space-y-3">
            <div className="flex items-baseline gap-2">
              <span className="font-mono text-3xl font-semibold text-foreground tabular">
                {clock.samples === 0 ? "—" : formatDuration(clock.skewMs, { signed: true })}
              </span>
              {clock.samples > 0 ? (
                <span className="font-mono text-xs text-faint-foreground tabular">
                  ~{formatDuration(clock.medianMs, { signed: true })}
                </span>
              ) : null}
            </div>
            <SkewMeter
              skewMs={clock.skewMs}
              warnMs={gateway.thresholds.warnMs}
              criticalMs={gateway.thresholds.criticalMs}
              samples={clock.samples}
            />
            <div className="grid grid-cols-2 gap-3 border-t border-border pt-3">
              <Stat
                label={dict.overview.drift}
                value={clock.driftDetermined ? `${formatDriftPerDay(clock.driftPerDayMs)}${dict.common.perDay}` : "—"}
                hint={clock.driftDetermined ? `R² ${clock.fit.toFixed(2)}` : dict.overview.driftHint}
                mono
                tone={Math.abs(clock.driftPerDayMs) > 1000 ? "warning" : "muted"}
              />
              <Stat
                label={dict.overview.roundTrip}
                value={clock.samples > 0 ? formatDuration(clock.roundTripMs) : "—"}
                mono
                tone="muted"
              />
            </div>
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader
            icon={<Gauge className="size-4" />}
            title={dict.gateway.healthPolicy}
            actions={
              <Badge
                tone={
                  health.state === "online"
                    ? "success"
                    : health.state === "degraded"
                      ? "warning"
                      : health.state === "offline"
                        ? "danger"
                        : "neutral"
                }
              >
                <StatusDot tone={healthTone} pulse={running && health.state === "online"} />
                {dict.states[health.state as keyof typeof dict.states] ?? dict.states.unknown}
              </Badge>
            }
          />
          <PanelBody className="space-y-3">
            <div className="flex items-baseline gap-2">
              <span className="font-mono text-3xl font-semibold text-foreground tabular">
                {health.window.length === 0 ? "—" : formatPercent(availability(health))}
              </span>
              <span className="text-xs text-faint-foreground">{dict.overview.availability}</span>
            </div>
            <OutcomeStrip window={health.window} />
            <div className="grid grid-cols-3 gap-3 border-t border-border pt-3">
              <Stat label="p50" value={latencies.length ? formatDuration(percentile(latencies, 0.5)) : "—"} mono tone="muted" />
              <Stat label="p95" value={latencies.length ? formatDuration(percentile(latencies, 0.95)) : "—"} mono tone="muted" />
              <Stat
                label={dict.gateway.failuresShort}
                value={health.consecutiveFailures}
                mono
                tone={health.consecutiveFailures > 0 ? "danger" : "muted"}
              />
            </div>
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader
            icon={<Waves className="size-4" />}
            title={dict.overview.eventRate}
            hint={dict.overview.eventRateHint}
          />
          <PanelBody className="space-y-3">
            <Sparkline values={frameRate} height={56} />
            <div className="grid grid-cols-2 gap-3 border-t border-border pt-3">
              <Stat
                label={dict.overview.nextPoll}
                value={mounted && running ? formatDuration(countdown) : dict.common.stopped}
                mono
                tone={running ? "primary" : "muted"}
              />
              <Stat
                label={dict.overview.lastCycle}
                value={lastCycleAt > 0 ? formatClock(lastCycleAt, false) : "—"}
                mono
                tone="muted"
              />
            </div>
          </PanelBody>
        </Panel>
      </div>

      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
        <Panel>
          <PanelHeader
            icon={<CircuitBoard className="size-4" />}
            title={dict.overview.emulator}
            actions={
              activeFaults > 0 ? (
                <Badge tone="warning">
                  {dict.emulator.activeFaults}: {activeFaults}
                </Badge>
              ) : (
                <Badge tone="success">{dict.emulator.presetHealthy}</Badge>
              )
            }
          />
          <PanelBody className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <Stat
                label={dict.overview.deviceTime}
                value={lastDeviceTime ? formatDeviceTime(new Date(lastDeviceTime)) : "—"}
                mono
              />
              <Stat
                label={dict.overview.serverTime}
                value={mounted ? formatDeviceTime(new Date(now)) : "—"}
                mono
                tone="muted"
              />
            </div>
            <div className="border-t border-border pt-2">
              <DefRow label={dict.overview.checksumMode} value={checksumMode} />
              <DefRow
                label={dict.emulator.clockOffset}
                value={faults.clockOffsetMs === 0 ? "0" : formatDuration(faults.clockOffsetMs, { signed: true })}
              />
              <DefRow
                label={dict.emulator.clockDrift}
                value={
                  faults.clockDriftPerDayMs === 0
                    ? "0"
                    : `${formatDriftPerDay(faults.clockDriftPerDayMs)}${dict.common.perDay}`
                }
              />
              <DefRow
                label={dict.emulator.responseDelay}
                value={faults.responseDelayMs === 0 ? "0" : formatDuration(faults.responseDelayMs)}
              />
            </div>
            {identityRead ? (
              <div className="grid grid-cols-3 gap-2 border-t border-border pt-3">
                <Stat label={dict.overview.model} value={identityRead.model} mono tone="muted" />
                <Stat label={dict.overview.serial} value={identityRead.serial} mono tone="muted" />
                <Stat label={dict.overview.firmware} value={identityRead.firmware} mono tone="muted" />
              </div>
            ) : null}
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader
            icon={<Router className="size-4" />}
            title={dict.overview.gateway}
            actions={
              <Badge tone={connected ? "success" : "neutral"}>
                <StatusDot tone={connected ? "online" : "idle"} pulse={running} />
                {connected ? dict.common.connected : dict.common.disconnected}
              </Badge>
            }
          />
          <PanelBody className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <Stat
                label={dict.overview.schedule}
                value={
                  gateway.scheduleMode === "aligned"
                    ? dict.gateway.scheduleAligned
                    : dict.gateway.scheduleInterval
                }
                hint={`${formatDuration(gateway.intervalMs)}${
                  gateway.scheduleMode === "aligned" && gateway.offsetMs > 0
                    ? ` +${formatDuration(gateway.offsetMs)}`
                    : ""
                }`}
              />
              <Stat
                label={dict.overview.successfulReads}
                value={counters.successfulReads}
                mono
                tone="success"
              />
            </div>
            <div className="border-t border-border pt-2">
              <DefRow label={dict.overview.failedReads} value={counters.failedReads} />
              <DefRow label={dict.overview.retries} value={counters.retries} />
              <DefRow label={dict.overview.protocolErrors} value={counters.protocolErrors} />
              <DefRow label={dict.overview.reconnects} value={counters.reconnects} />
            </div>
            <div className="border-t border-border pt-3">
              <p className="eyebrow pb-1.5">{dict.gateway.skewChart}</p>
              <SkewHistory
                samples={samples}
                medianMs={clock.medianMs}
                warnMs={gateway.thresholds.warnMs}
              />
            </div>
          </PanelBody>
        </Panel>
      </div>
    </div>
  );
}
