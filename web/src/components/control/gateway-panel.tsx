"use client";

import { CalendarClock, Lock, Repeat, Router, ShieldAlert } from "lucide-react";
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
import { useTelemetry } from "@/lib/telemetry/use-telemetry";
import { useFormat } from "@/lib/use-format";
import { useNow } from "@/lib/use-now";
import { useBenchStore } from "@/stores/bench-store";

/**
 * Gateway control.
 *
 * The schedule section is the one that needs showing rather than describing:
 * the difference between "every 5 seconds" and "on the fifth second of every
 * minute" is invisible in a form and obvious on a timeline, so the ticks sit
 * directly under the control that produces them.
 *
 * ## Editable on the bench, read-only on the live stack
 *
 * Interval, thresholds and the availability policy are the gateway process's
 * own configuration, and its control plane has no setter for any of them. So
 * on the live source this panel becomes a faithful description of how the
 * running gateway is configured, with every control replaced by its value
 * rather than disabled — a greyed-out `<select>` can only display a value that
 * happens to be one of its options, and a gateway set to an interval this UI
 * never offers would render blank.
 */

const INTERVALS = [1000, 2000, 5000, 10_000, 30_000, 60_000];
const OFFSETS = [0, 1000, 2000, 5000, 10_000, 15_000, 30_000];
const TIMEOUTS = [250, 500, 1000, 1500, 3000];
const RETRIES = [0, 1, 2, 3, 5];
const WARN_THRESHOLDS = [500, 1000, 2000, 5000, 10_000];
const CRITICAL_THRESHOLDS = [5000, 10_000, 30_000, 60_000, 300_000];
const DEGRADE_AFTER = [1, 2, 3, 5, 8];
const OFFLINE_AFTER = [3, 5, 10, 15, 20];
const RECOVER_AFTER = [1, 2, 3, 5];

export function GatewayPanel() {
  const dict = useDictionary();
  const format = useFormat();

  const telemetry = useTelemetry();
  const {
    editable,
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
  } = telemetry;

  // Only the bench has settings to write; on the live source the setter is
  // never reached, because every control is replaced by a readout.
  const patchGateway = useBenchStore((state) => state.patchGateway);

  const now = useNow();

  const alignedHint =
    schedule.mode === "aligned"
      ? dict.gateway.scheduleAlignedHint
      : dict.gateway.scheduleIntervalHint;

  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_23rem]">
      <div className="space-y-4">
        <SourceNotice />

        {editable ? null : (
          <p className="flex items-start gap-2 rounded-md border border-border bg-surface-raised/40 px-3 py-2 text-xs leading-relaxed text-faint-foreground">
            <Lock className="mt-0.5 size-3.5 shrink-0" />
            {dict.source.readOnlyHint}
          </p>
        )}

        <Panel>
          <PanelHeader
            icon={<CalendarClock className="size-4" />}
            title={dict.gateway.title}
            hint={dict.gateway.subtitle}
            actions={
              <>
                {editable ? null : <Badge tone="neutral">{dict.source.readOnly}</Badge>}
                <Badge tone={running ? "success" : "neutral"}>
                  <StatusDot tone={running ? "online" : "idle"} pulse={running} />
                  {running ? dict.gateway.pollingRunning : dict.gateway.pollingStopped}
                </Badge>
              </>
            }
          />
          <PanelBody className="space-y-3.5">
            <Field label={dict.gateway.scheduleMode} hint={alignedHint}>
              {editable ? (
                <SegmentedControl<ScheduleMode>
                  value={schedule.mode}
                  onChange={(mode) => patchGateway({ scheduleMode: mode })}
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
                editable={editable}
                value={schedule.intervalMs}
                display={format.duration(schedule.intervalMs)}
                options={INTERVALS}
                render={(value) => format.duration(value)}
                onChange={(value) => patchGateway({ intervalMs: value })}
              />
              <Setting
                label={dict.gateway.offset}
                hint={schedule.mode === "interval" ? dict.gateway.scheduleIntervalHint : undefined}
                editable={editable}
                disabled={schedule.mode !== "aligned"}
                value={schedule.offsetMs}
                display={schedule.offsetMs === 0 ? "0" : format.duration(schedule.offsetMs)}
                options={OFFSETS.filter((value) => value < schedule.intervalMs)}
                render={(value) => (value === 0 ? "0" : format.duration(value))}
                onChange={(value) => patchGateway({ offsetMs: value })}
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
          <PanelBody className="grid gap-3 sm:grid-cols-2">
            <Setting
              label={dict.gateway.requestTimeout}
              editable={editable}
              value={limits.requestTimeoutMs}
              display={format.duration(limits.requestTimeoutMs)}
              options={TIMEOUTS}
              render={(value) => format.duration(value)}
              onChange={(value) => patchGateway({ requestTimeoutMs: value })}
            />
            <Setting
              label={dict.gateway.retryAttempts}
              editable={editable}
              value={limits.retryAttempts}
              display={String(limits.retryAttempts)}
              options={RETRIES}
              render={String}
              onChange={(value) => patchGateway({ retryAttempts: value })}
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
                editable={editable}
                value={thresholds.warnMs}
                display={format.duration(thresholds.warnMs)}
                options={WARN_THRESHOLDS}
                render={(value) => format.duration(value)}
                onChange={(value) =>
                  patchGateway({ thresholds: { ...thresholds, warnMs: value } })
                }
              />
              <Setting
                label={dict.gateway.criticalThreshold}
                editable={editable}
                value={thresholds.criticalMs}
                display={format.duration(thresholds.criticalMs)}
                options={CRITICAL_THRESHOLDS}
                render={(value) => format.duration(value)}
                onChange={(value) =>
                  patchGateway({ thresholds: { ...thresholds, criticalMs: value } })
                }
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
            {editable ? null : (
              <p className="text-xs text-faint-foreground">{dict.source.noReset}</p>
            )}
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader title={dict.gateway.healthPolicy} hint={dict.gateway.healthPolicyHint} />
          <PanelBody className="space-y-3">
            <div className="grid grid-cols-3 gap-2">
              <Setting
                label={dict.gateway.degradeAfter}
                editable={editable}
                value={policy.degradeAfter}
                display={String(policy.degradeAfter)}
                options={DEGRADE_AFTER}
                render={String}
                onChange={(value) => patchGateway({ policy: { ...policy, degradeAfter: value } })}
              />
              <Setting
                label={dict.gateway.offlineAfter}
                editable={editable}
                value={policy.offlineAfter}
                display={String(policy.offlineAfter)}
                options={OFFLINE_AFTER}
                render={String}
                onChange={(value) => patchGateway({ policy: { ...policy, offlineAfter: value } })}
              />
              <Setting
                label={dict.gateway.recoverAfter}
                editable={editable}
                value={policy.recoverAfter}
                display={String(policy.recoverAfter)}
                options={RECOVER_AFTER}
                render={String}
                onChange={(value) => patchGateway({ policy: { ...policy, recoverAfter: value } })}
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
 * The two forms are bound together here rather than at each of the nine call
 * sites, so a control cannot end up editable on a source that will silently
 * discard the change.
 */
function Setting({
  label,
  hint,
  editable,
  disabled,
  value,
  display,
  options,
  render,
  onChange,
}: {
  label: ReactNode;
  hint?: ReactNode;
  editable: boolean;
  disabled?: boolean;
  value: number;
  display: string;
  options: number[];
  render: (value: number) => string;
  onChange: (value: number) => void;
}) {
  return (
    <Field label={label} hint={hint}>
      {editable ? (
        <Select
          value={String(value)}
          disabled={disabled}
          onChange={(event) => onChange(Number(event.target.value))}
        >
          {options.map((option) => (
            <option key={option} value={option}>
              {render(option)}
            </option>
          ))}
        </Select>
      ) : (
        <ReadonlyValue>{display}</ReadonlyValue>
      )}
    </Field>
  );
}
