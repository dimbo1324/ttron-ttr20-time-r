"use client";

import { Server } from "lucide-react";

import { useDictionary } from "@/components/locale-provider";
import { Badge } from "@/components/ui/badge";
import { Panel, PanelBody, PanelHeader } from "@/components/ui/panel";
import { StateBadge } from "@/components/ui/state-badge";
import { StatusDot } from "@/components/ui/status-dot";
import { CLOCK_TONE } from "@/lib/status";
import type { FleetView } from "@/lib/telemetry/types";
import { useFormat } from "@/lib/use-format";

/**
 * Every device the gateway is polling.
 *
 * A fleet is read by scanning down a column, not by reading rows, so the two
 * columns that decide whether a device needs attention — is it answering, and
 * is its clock sane — are adjacent and coloured, and the skew beside them is
 * signed so a meter running fast is distinguishable from one running slow at a
 * glance.
 *
 * The worst clock in the fleet is called out in the header rather than left to
 * be found: on a bench with three devices sorting by eye is free, and on a
 * real installation with forty it is not.
 */
export function FleetPanel({ fleet }: { fleet: FleetView }) {
  const dict = useDictionary();
  const format = useFormat();

  return (
    <Panel>
      <PanelHeader
        icon={<Server className="size-4" />}
        title={dict.fleet.title}
        hint={dict.fleet.hint}
        actions={
          fleet.worstDeviceId ? (
            <Badge tone={fleet.clockCritical > 0 ? "danger" : fleet.clockWarn > 0 ? "warning" : "neutral"}>
              {dict.fleet.worst}
              <span className="font-mono tabular">
                {fleet.worstDeviceId} {format.duration(fleet.worstSkewMs, { signed: true })}
              </span>
            </Badge>
          ) : null
        }
      />
      <PanelBody className="p-0">
        {fleet.devices.length === 0 ? (
          <p className="p-3.5 text-xs text-faint-foreground">{dict.fleet.empty}</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full border-collapse">
              <thead>
                <tr className="border-b border-border text-left">
                  <Th>{dict.fleet.device}</Th>
                  <Th>{dict.fleet.target}</Th>
                  <Th>{dict.fleet.state}</Th>
                  <Th>{dict.fleet.clock}</Th>
                  <Th className="text-right">{dict.fleet.skew}</Th>
                  <Th className="text-right">{dict.fleet.availability}</Th>
                  <Th className="text-right">{dict.fleet.samples}</Th>
                </tr>
              </thead>
              <tbody>
                {fleet.devices.map((device) => (
                  <tr key={device.id} className="border-b border-border/50 last:border-0">
                    <td className="px-3.5 py-1.5 align-middle">
                      <span className="flex items-center gap-1.5">
                        <StatusDot
                          tone={device.running ? "online" : "idle"}
                          pulse={device.running}
                        />
                        <span className="truncate text-xs text-foreground">{device.name}</span>
                      </span>
                    </td>
                    <td className="py-1.5 align-middle font-mono text-[0.6875rem] text-muted-foreground tabular">
                      {device.target}
                    </td>
                    <td className="py-1.5 align-middle">
                      <StateBadge kind="health" state={device.health} />
                    </td>
                    <td className="py-1.5 align-middle">
                      <StateBadge kind="clock" state={device.clock} />
                    </td>
                    <td
                      className={`py-1.5 text-right align-middle font-mono text-xs tabular ${
                        TONE_TEXT[CLOCK_TONE[device.clock]]
                      }`}
                    >
                      {device.samples === 0
                        ? dict.common.none
                        : format.duration(device.skewMs, { signed: true })}
                    </td>
                    <td className="py-1.5 text-right align-middle font-mono text-xs text-muted-foreground tabular">
                      {format.percent(device.availability)}
                    </td>
                    <td className="py-1.5 pr-3.5 text-right align-middle font-mono text-xs text-faint-foreground tabular">
                      {format.count(device.samples)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {fleet.devices.length === 1 ? (
          <p className="border-t border-border px-3.5 py-2 text-xs text-faint-foreground">
            {dict.fleet.single}
          </p>
        ) : null}
      </PanelBody>
    </Panel>
  );
}

/** The badge tones, as text colours, so the skew figure matches its own state. */
const TONE_TEXT = {
  neutral: "text-muted-foreground",
  success: "text-foreground",
  warning: "text-warning",
  danger: "text-destructive",
} as const;

function Th({ children, className = "" }: { children: React.ReactNode; className?: string }) {
  return (
    <th
      scope="col"
      className={`px-3.5 py-1.5 text-[0.625rem] font-medium tracking-wide text-faint-foreground uppercase ${className}`}
    >
      {children}
    </th>
  );
}
