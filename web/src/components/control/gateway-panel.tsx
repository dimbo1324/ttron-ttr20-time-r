"use client";

import { CalendarClock, Lock, Repeat, Router, ShieldAlert, TriangleAlert } from "lucide-react";
import type { ReactNode } from "react";

import { ScheduleTimeline, SkewHistory } from "@/components/dashboard/charts";
import { SourceNotice } from "@/components/layout/source-notice";
import { useDictionary } from "@/components/locale-provider";
import { Badge } from "@/components/ui/badge";
import { Field, ReadonlyValue, SegmentedControl, Select } from "@/components/ui/controls";
import { DefRow, Panel, PanelBody, PanelHeader, Stat } from "@/components/ui/panel";
import { StateBadge } from "@/components/ui/state-badge";
import { StatusDot } from "@/components/ui/status-dot";
import type { ScheduleMode } from "@/lib/bench/domain";
import { useSettingsControls, useTelemetry } from "@/lib/telemetry/use-telemetry";
import { useFormat } from "@/lib/use-format";
import { useNow } from "@/lib/use-now";

/**
 * Gateway control.
 *
 * The schedule section is the one that needs showing rather than describing:
 * the difference between "every 5 seconds" and "on the fifth second of every
 * minute" is invisible in a form and obvious on a timeline, so the ticks sit
 * directly under the control that produces them.
 *
 * ## The same controls, two gateways behind them
 *
 * On the bench these settings drive the engine in this tab. On the live source
 * they reconfigure the running Go gateway over the API, and every change lands
 * on a process that may be mid-poll -- the gateway validates the whole
 * configuration before applying any of it, so a rejected value leaves it
 * exactly as it was and this panel simply keeps showing the truth.
 *
 * A control is replaced by its value, never disabled, when it cannot be
 * written: a greyed-out `<select>` can only display a value that happens to be
 * one of its options, and a gateway configured with an interval this UI never
 * offers would render blank -- which reads as "not set" rather than "not
 * editable".
 */

const INTERVALS = [1000, 2000, 5000, 10_000, 30_000, 60_000];
const OFFSETS = [0, 1000, 2000, 5000, 10_000, 15_000, 30_000];
const TIMEOUTS = [250, 500, 1000, 1500, 3000];
const RETRIES = [0, 1, 2, 3, 5];
const RETRY_DELAYS = [50, 100, 200, 500, 1000];
const WARN_THRESHOLDS = [500, 1000, 2000, 5000, 10_000];
const CRITICAL_THRESHOLDS = [5000, 10_000, 30_000, 60_000, 300_000];
const DEGRADE_AFTER = [1, 2, 3, 5, 8];
const OFFLINE_AFTER = [3, 5, 10, 15, 20];
const RECOVER_AFTER = [1, 2, 3, 5];

/**
 * The offered values, plus whatever the gateway is actually set to.
 *
 * A gateway configured outside this list is not a bug to hide: an operator
 * needs to see the real 7-second interval and be able to leave it alone while
 * changing something else. Without this the select would render blank and the
 * first edit would silently snap the value to whichever option came first.
 */
function withCurrent(options: number[], current: number): number[] {
  if (options.includes(current)) return options;
  return [...options, current].sort((left, right) => left - right);
}

export function GatewayPanel() {
  const dict = useDictionary();
  const format = useFormat();

  const telemetry = useTelemetry();
  const {
    source,
    editable,
    deviceName,
    running,
    connected,
    counters,
    health,
    samples,
    ticks,
    schedule,
    thresholds,
    policy,
    clock,
    limits,
    target,
    settingsError,
  } = telemetry;

  const { setSettings, busy } = useSettingsControls();
  const now = useNow();

  // A write in flight leaves every control frozen rather than only the one
  // that was touched: the gateway takes the whole configuration at once, so a
  // second edit sent mid-flight would race the first.
  const writable = editable && !busy;

  const alignedHint =
    schedule.mode === "aligned"
      ? dict.gateway.scheduleAlignedHint
      : dict.gateway.scheduleIntervalHint;

  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_23rem]">
      <div className="space-y-4">
        <SourceNotice />

        {source === "live" ? (
          <p className="flex items-start gap-2 rounded-md border border-border bg-surface-raised/40 px-3 py-2 text-xs leading-relaxed text-faint-foreground">
            <Lock className="mt-0.5 size-3.5 shrink-0" />
            <span>
              {editable ? dict.source.liveSettingsHint : dict.source.settingsUnavailable}
              {deviceName ? (
                <>
                  {" "}
                  <span className="text-muted-foreground">
                    {dict.source.controlledDevice}: <span className="font-mono">{deviceName}</span>
                  </span>
                </>
              ) : null}
            </span>
          </p>
        ) : null}

        {settingsError ? (
          <p
            role="status"
            className="flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/8 px-3 py-2 text-xs leading-relaxed"
          >
            <TriangleAlert className="mt-0.5 size-3.5 shrink-0 text-destructive" />
            <span>
              <span className="font-medium text-foreground">{dict.source.settingsRejected}</span>{" "}
              <span className="font-mono text-muted-foreground">{settingsError}</span>
            </span>
          </p>
        ) : null}

        <Panel>
          <PanelHeader
            icon={<CalendarClock className="size-4" />}
            title={dict.gateway.title}
            hint={dict.gateway.subtitle}
            actions={
              <>
                {source === "live" && !editable ? (
                  <Badge tone="neutral">{dict.source.readOnly}</Badge>
                ) : null}
                <Badge tone={running ? "success" : "neutral"}>
                  <StatusDot tone={running ? "online" : "idle"} pulse={running} />
                  {running ? dict.gateway.pollingRunning : dict.gateway.pollingStopped}
                </Badge>
              </>
            }
          />
          <PanelBody className="space-y-3.5">
            <Field label={dict.gateway.scheduleMode} hint={alignedHint}>
              {writable ? (
                <SegmentedControl<ScheduleMode>
                  value={schedule.mode}
                  onChange={(mode) =>
                    setSettings(
                      // Switching to interval drops the offset, which only an
                      // aligned schedule has a use for -- and which the gateway
                      // rejects when it is not below the interval.
                      mode === "aligned"
                        ? { scheduleMode: mode }
                        : { scheduleMode: mode, offsetMs: 0 },
                    )
                  }
                  options={[
                    { value: "interval", label: dict.gateway.scheduleInterval },
                    { value: "aligned", label: dict.gateway.scheduleAligned },
                  ]}
                  className="w-full"
                />
              ) : (
                <ReadonlyValue>
                  {schedule.mode === "aligned"
                    ? dict.gateway.scheduleAligned
                    : dict.gateway.scheduleInterval}
                </ReadonlyValue>
              )}
            </Field>

            <div className="grid gap-3 sm:grid-cols-2">
              <Setting
                label={dict.gateway.interval}
                writable={writable}
                value={schedule.intervalMs}
                options={withCurrent(INTERVALS, schedule.intervalMs)}
                render={(value) => format.duration(value)}
                // Every interval here is reachable: shortening one pulls the
                // offset and the request timeout into range in the same patch,
                // rather than sending a combination the gateway would refuse.
                onChange={(value) => setSettings({ intervalMs: value })}
              />
              <Setting
                label={dict.gateway.offset}
                hint={schedule.mode === "interval" ? dict.gateway.scheduleIntervalHint : undefined}
                writable={writable && schedule.mode === "aligned"}
                value={schedule.offsetMs}
                options={withCurrent(
                  OFFSETS.filter((value) => value < schedule.intervalMs),
                  schedule.offsetMs,
                )}
                render={(value) => (value === 0 ? "0" : format.duration(value))}
                onChange={(value) => setSettings({ offsetMs: value })}
              />
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
          <PanelBody className="grid gap-3 sm:grid-cols-3">
            <Setting
              label={dict.gateway.requestTimeout}
              writable={writable}
              value={limits.requestTimeoutMs}
              options={withCurrent(
                // A request that may outlast its own interval is not a
                // schedule, and the gateway rejects it; the list never offers
                // a value that would be refused.
                TIMEOUTS.filter((value) => value < schedule.intervalMs),
                limits.requestTimeoutMs,
              )}
              render={(value) => format.duration(value)}
              onChange={(value) => setSettings({ requestTimeoutMs: value })}
            />
            <Setting
              label={dict.gateway.retryAttempts}
              writable={writable}
              value={limits.retryAttempts}
              options={withCurrent(RETRIES, limits.retryAttempts)}
              render={String}
              onChange={(value) => setSettings({ retryAttempts: value })}
            />
            <Setting
              label={dict.gateway.retryDelay}
              hint={dict.gateway.retryDelayHint}
              writable={writable}
              value={limits.retryDelayMs}
              options={withCurrent(RETRY_DELAYS, limits.retryDelayMs)}
              render={(value) => format.duration(value)}
              onChange={(value) => setSettings({ retryDelayMs: value })}
            />
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
              <Setting
                label={dict.gateway.warnThreshold}
                writable={writable}
                value={thresholds.warnMs}
                options={withCurrent(WARN_THRESHOLDS, thresholds.warnMs)}
                render={(value) => format.duration(value)}
                onChange={(value) => setSettings({ warnMs: value })}
              />
              <Setting
                label={dict.gateway.criticalThreshold}
                writable={writable}
                value={thresholds.criticalMs}
                options={withCurrent(
                  // A critical threshold below the warning one is not an
                  // escalation, and the gateway refuses it.
                  CRITICAL_THRESHOLDS.filter((value) => value >= thresholds.warnMs),
                  thresholds.criticalMs,
                )}
                render={(value) => format.duration(value)}
                onChange={(value) => setSettings({ criticalMs: value })}
              />
            </div>
            <div>
              <p className="eyebrow pb-1.5">{dict.gateway.skewChart}</p>
              <SkewHistory
                samples={samples}
                medianMs={clock.medianMs}
                warnMs={thresholds.warnMs}
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
                value={
                  now && running && schedule.nextPollAt
                    ? format.duration(Math.max(0, schedule.nextPollAt - now))
                    : dict.common.stopped
                }
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
              {target ? <DefRow label={dict.source.target} value={target} /> : null}
              <DefRow label={dict.overview.successfulReads} value={counters.successfulReads} />
              <DefRow label={dict.overview.failedReads} value={counters.failedReads} />
              <DefRow label={dict.overview.retries} value={counters.retries} />
              <DefRow label={dict.overview.protocolErrors} value={counters.protocolErrors} />
              <DefRow label={dict.overview.reconnects} value={counters.reconnects} />
              <DefRow label={dict.overview.sessions} value={counters.connections} />
            </div>
            {source === "live" ? (
              <p className="text-xs text-faint-foreground">{dict.source.noReset}</p>
            ) : null}
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader title={dict.gateway.healthPolicy} hint={dict.gateway.healthPolicyHint} />
          <PanelBody className="space-y-3">
            <div className="grid grid-cols-3 gap-2">
              <Setting
                label={dict.gateway.degradeAfter}
                writable={writable}
                value={policy.degradeAfter}
                options={withCurrent(
                  DEGRADE_AFTER.filter((value) => value <= policy.offlineAfter),
                  policy.degradeAfter,
                )}
                render={String}
                onChange={(value) => setSettings({ degradeAfter: value })}
              />
              <Setting
                label={dict.gateway.offlineAfter}
                writable={writable}
                value={policy.offlineAfter}
                options={withCurrent(
                  // Offline before degraded is not an escalation either.
                  OFFLINE_AFTER.filter((value) => value >= policy.degradeAfter),
                  policy.offlineAfter,
                )}
                render={String}
                onChange={(value) => setSettings({ offlineAfter: value })}
              />
              <Setting
                label={dict.gateway.recoverAfter}
                writable={writable}
                value={policy.recoverAfter}
                options={withCurrent(RECOVER_AFTER, policy.recoverAfter)}
                render={String}
                onChange={(value) => setSettings({ recoverAfter: value })}
              />
            </div>
            <div className="border-t border-border pt-2">
              <DefRow label={dict.gateway.failuresShort} value={health.consecutiveFailures} />
              <DefRow label={dict.gateway.successesShort} value={health.consecutiveSuccesses} />
              <DefRow
                label={dict.gateway.deviceState}
                value={<StateBadge kind="health" state={health.state} />}
              />
            </div>
          </PanelBody>
        </Panel>
      </div>
    </div>
  );
}

/**
 * One setting: a picker where it can be changed, its value where it cannot.
 *
 * The two forms are bound together here rather than at each of the ten call
 * sites, so a control cannot end up editable on a source that would discard
 * the change.
 */
function Setting({
  label,
  hint,
  writable,
  value,
  options,
  render,
  onChange,
}: {
  label: ReactNode;
  hint?: ReactNode;
  writable: boolean;
  value: number;
  options: number[];
  render: (value: number) => string;
  onChange: (value: number) => void;
}) {
  return (
    <Field label={label} hint={hint}>
      {writable ? (
        <Select value={String(value)} onChange={(event) => onChange(Number(event.target.value))}>
          {options.map((option) => (
            <option key={option} value={option}>
              {render(option)}
            </option>
          ))}
        </Select>
      ) : (
        <ReadonlyValue>{render(value)}</ReadonlyValue>
      )}
    </Field>
  );
}

