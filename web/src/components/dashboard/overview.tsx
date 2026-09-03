"use client";

import { Clock, CircuitBoard, Gauge, Router, Waves } from "lucide-react";
import { useMemo } from "react";

import { useDictionary } from "@/components/locale-provider";
import { OutcomeStrip, SkewHistory, SkewMeter, Sparkline } from "@/components/dashboard/charts";
import { FleetPanel } from "@/components/dashboard/fleet";
import { SourceNotice } from "@/components/layout/source-notice";
import { Badge } from "@/components/ui/badge";
import { Panel, PanelBody, PanelHeader, DefRow, Stat } from "@/components/ui/panel";
import { StateBadge } from "@/components/ui/state-badge";
import { StatusDot } from "@/components/ui/status-dot";
import { formatDeviceTime } from "@/lib/ft12";
import { activeFaultCount } from "@/lib/telemetry/types";
import { useTelemetry } from "@/lib/telemetry/use-telemetry";
import { useFormat } from "@/lib/use-format";
import { useNow } from "@/lib/use-now";

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
  const format = useFormat();

  const {
    running,
    connected,
    events,
    samples,
    health,
    counters,
    faults,
    fleet,
    schedule,
    thresholds,
    checksumMode,
    target,
    lastDeviceTime,
    identity,
    lastCycleAt,
    clock,
  } = useTelemetry();

  const activeFaults = activeFaultCount(faults);
  const now = useNow();

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

  const countdown = Math.max(0, schedule.nextPollAt - now);

  return (
    <div className="space-y-4">
      <SourceNotice />
      <div className="grid gap-4 lg:grid-cols-3">
        <Panel>
          <PanelHeader
            icon={<Clock className="size-4" />}
            title={dict.overview.skew}
            hint={dict.overview.skewHint}
            actions={<StateBadge kind="clock" state={clock.state} />}
          />
          <PanelBody className="space-y-3">
            <div className="flex items-baseline gap-2">
              <span className="font-mono text-3xl font-semibold text-foreground tabular">
                {clock.samples === 0 ? "—" : format.duration(clock.skewMs, { signed: true })}
              </span>
              {clock.samples > 0 ? (
                <span className="font-mono text-xs text-faint-foreground tabular">
                  ~{format.duration(clock.medianMs, { signed: true })}
                </span>
              ) : null}
            </div>
            <SkewMeter
              skewMs={clock.skewMs}
              warnMs={thresholds.warnMs}
              criticalMs={thresholds.criticalMs}
              samples={clock.samples}
            />
            <div className="grid grid-cols-2 gap-3 border-t border-border pt-3">
              <Stat
                label={dict.overview.drift}
                value={clock.driftDetermined ? format.driftPerDay(clock.driftPerDayMs) : "—"}
                hint={clock.driftDetermined ? `R² ${clock.fit.toFixed(2)}` : dict.overview.driftHint}
                mono
                tone={Math.abs(clock.driftPerDayMs) > 1000 ? "warning" : "muted"}
              />
              <Stat
                label={dict.overview.roundTrip}
                value={clock.samples > 0 ? format.duration(clock.roundTripMs) : "—"}
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
              <StateBadge
                kind="health"
                state={health.state}
                pulse={running && health.state === "online"}
              />
            }
          />
          <PanelBody className="space-y-3">
            <div className="flex items-baseline gap-2">
              <span className="font-mono text-3xl font-semibold text-foreground tabular">
                {health.windowSamples === 0 ? "—" : format.percent(health.availability)}
              </span>
              <span className="text-xs text-faint-foreground">{dict.overview.availability}</span>
            </div>
            <OutcomeStrip window={health.window} />
            <div className="grid grid-cols-3 gap-3 border-t border-border pt-3">
              <Stat
                label="p50"
                value={health.windowSamples ? format.duration(health.latencyP50Ms) : "—"}
                mono
                tone="muted"
              />
              <Stat
                label="p95"
                value={health.windowSamples ? format.duration(health.latencyP95Ms) : "—"}
                mono
                tone="muted"
              />
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
            <Sparkline values={frameRate} height={56} label={dict.overview.eventRateHint} />
            <div className="grid grid-cols-2 gap-3 border-t border-border pt-3">
              <Stat
                label={dict.overview.nextPoll}
                value={now && running ? format.duration(countdown) : dict.common.stopped}
                mono
                tone={running ? "primary" : "muted"}
              />
              <Stat
                label={dict.overview.lastCycle}
                value={lastCycleAt > 0 ? format.clock(lastCycleAt, false) : "—"}
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
                value={now ? formatDeviceTime(new Date(now)) : dict.common.none}
                mono
                tone="muted"
              />
            </div>
            <div className="border-t border-border pt-2">
              <DefRow label={dict.overview.checksumMode} value={checksumMode} />
              {target ? <DefRow label={dict.source.target} value={target} /> : null}
              {/* The clock rows are dropped rather than zeroed when the source
                  cannot move the device clock: a "0" would claim it had been
                  checked and found correct. */}
              {faults && faults.clockOffsetMs !== null ? (
                <DefRow
                  label={dict.emulator.clockOffset}
                  value={
                    faults.clockOffsetMs === 0
                      ? "0"
                      : format.duration(faults.clockOffsetMs, { signed: true })
                  }
                />
              ) : null}
              {faults && faults.clockDriftPerDayMs !== null ? (
                <DefRow
                  label={dict.emulator.clockDrift}
                  value={
                    faults.clockDriftPerDayMs === 0
                      ? "0"
                      : format.driftPerDay(faults.clockDriftPerDayMs)
                  }
                />
              ) : null}
              {faults ? (
                <DefRow
                  label={dict.emulator.responseDelay}
                  value={faults.responseDelayMs === 0 ? "0" : format.duration(faults.responseDelayMs)}
                />
              ) : null}
            </div>
            {identity ? (
              <div className="grid grid-cols-3 gap-2 border-t border-border pt-3">
                <Stat label={dict.overview.model} value={identity.model} mono tone="muted" />
                <Stat label={dict.overview.serial} value={identity.serial} mono tone="muted" />
                <Stat label={dict.overview.firmware} value={identity.firmware} mono tone="muted" />
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
                  schedule.mode === "aligned"
                    ? dict.gateway.scheduleAligned
                    : dict.gateway.scheduleInterval
                }
                hint={`${format.duration(schedule.intervalMs)}${
                  schedule.mode === "aligned" && schedule.offsetMs > 0
                    ? ` +${format.duration(schedule.offsetMs)}`
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
              <SkewHistory samples={samples} medianMs={clock.medianMs} warnMs={thresholds.warnMs} />
            </div>
          </PanelBody>
        </Panel>
      </div>

      {/* Only a source that polls more than its own device has a fleet. */}
      {fleet ? <FleetPanel fleet={fleet} /> : null}
    </div>
  );
}
