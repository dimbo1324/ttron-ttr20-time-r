"use client";

import { ArrowLeftRight, Radio, Trash2 } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";

import { SourceNotice } from "@/components/layout/source-notice";
import { useDictionary } from "@/components/locale-provider";
import { FrameInspector } from "@/components/protocol/frame-view";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input, SegmentedControl } from "@/components/ui/controls";
import { Panel, PanelBody, PanelHeader, Stat } from "@/components/ui/panel";
import { formatHex } from "@/lib/ft12";
import { DIRECTION_BG, DIRECTION_TEXT, DIRECTION_TONE } from "@/lib/status";
import type { EventDirection, LogEvent } from "@/lib/telemetry/types";
import { useTelemetry } from "@/lib/telemetry/use-telemetry";
import { useFormat } from "@/lib/use-format";
import { cn } from "@/lib/utils";
import { useBenchStore } from "@/stores/bench-store";

/**
 * Exchange monitor.
 *
 * The log is the primary object, but a log alone does not show *timing*, and
 * timing is where protocol faults live — a retry looks identical to a normal
 * poll in a list of rows and completely different on a time axis. So the log
 * sits under a sequence diagram that draws the same events against real
 * milliseconds, and selecting a row decodes it in full beside both.
 */

type Filter = "all" | EventDirection;

export function ExchangeMonitor() {
  const dict = useDictionary();
  const { source, events, checksumMode, counters } = useTelemetry();
  /**
   * Clearing is bench-only: the live log is the gateway's own frame history,
   * and a console must not be able to erase a device's record from a toolbar.
   */
  const clearBenchEvents = useBenchStore((state) => state.clearEvents);
  const clearEvents = source === "bench" ? clearBenchEvents : null;

  const [filter, setFilter] = useState<Filter>("all");
  const [query, setQuery] = useState("");
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [autoscroll, setAutoscroll] = useState(true);
  const listRef = useRef<HTMLDivElement>(null);

  const filtered = useMemo(() => {
    const needle = query.trim().toLowerCase();
    return events.filter((event) => {
      if (filter !== "all" && event.direction !== filter) return false;
      if (!needle) return true;
      return (
        event.command.toLowerCase().includes(needle) ||
        formatHex(event.bytes).toLowerCase().includes(needle)
      );
    });
  }, [events, filter, query]);

  useEffect(() => {
    if (!autoscroll || !listRef.current) return;
    listRef.current.scrollTop = listRef.current.scrollHeight;
  }, [filtered.length, autoscroll]);

  const selected = useMemo(
    () => events.find((event) => event.id === selectedId) ?? null,
    [events, selectedId],
  );

  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_23rem]">
      <div className="space-y-4">
        <SourceNotice />
        <Panel>
          <PanelHeader
            icon={<ArrowLeftRight className="size-4" />}
            title={dict.monitor.exchange}
            hint={dict.monitor.exchangeHint}
          />
          <PanelBody>
            <SequenceDiagram events={events} />
          </PanelBody>
        </Panel>

        <Panel className="min-h-0">
          <PanelHeader
            icon={<Radio className="size-4" />}
            title={dict.monitor.title}
            hint={dict.monitor.subtitle}
            actions={
              <>
                <Button
                  size="xs"
                  variant={autoscroll ? "subtle" : "ghost"}
                  onClick={() => setAutoscroll((value) => !value)}
                >
                  {dict.monitor.autoscroll}
                </Button>
                {clearEvents ? (
                  <Button
                    size="iconSm"
                    variant="ghost"
                    onClick={clearEvents}
                    title={dict.monitor.clearLog}
                  >
                    <Trash2 className="size-4" />
                  </Button>
                ) : null}
              </>
            }
          />

          <div className="flex flex-wrap items-center gap-2 border-b border-border px-3.5 py-2">
            <SegmentedControl<Filter>
              value={filter}
              onChange={setFilter}
              options={[
                { value: "all", label: dict.common.total },
                { value: "tx", label: dict.directions.tx },
                { value: "rx", label: dict.directions.rx },
                { value: "err", label: dict.directions.err },
                { value: "sys", label: dict.directions.system },
              ]}
            />
            <Input
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder={dict.monitor.search}
              className="h-8 max-w-[16rem] flex-1 font-mono text-xs"
            />
            <span className="ml-auto font-mono text-[0.6875rem] text-faint-foreground tabular">
              {filtered.length} {dict.monitor.frames}
            </span>
          </div>

          <div ref={listRef} className="max-h-[26rem] min-h-[12rem] overflow-y-auto">
            {filtered.length === 0 ? (
              <div className="flex h-40 flex-col items-center justify-center gap-1 text-center">
                <p className="text-sm text-muted-foreground">{dict.monitor.empty}</p>
                <p className="text-xs text-faint-foreground">{dict.monitor.emptyHint}</p>
              </div>
            ) : (
              <table className="w-full border-collapse">
                <thead className="sticky top-0 z-10 bg-card">
                  <tr className="border-b border-border text-left">
                    <th scope="col" className="w-[5.5rem] px-3.5 py-1 text-[0.625rem] font-medium tracking-wide text-faint-foreground uppercase">
                      {dict.monitor.time}
                    </th>
                    <th
                      scope="col"
                      title={dict.monitor.direction}
                      className="w-11 py-1 text-[0.625rem] font-medium tracking-wide text-faint-foreground uppercase"
                    >
                      <abbr title={dict.monitor.direction} className="no-underline">
                        {dict.monitor.dir}
                      </abbr>
                    </th>
                    <th scope="col" className="w-[7.5rem] py-1 text-[0.625rem] font-medium tracking-wide text-faint-foreground uppercase">
                      {dict.monitor.commandColumn}
                    </th>
                    <th scope="col" className="py-1 text-[0.625rem] font-medium tracking-wide text-faint-foreground uppercase">
                      {dict.monitor.raw}
                    </th>
                    <th scope="col" className="w-14 py-1 pr-3.5 text-right text-[0.625rem] font-medium tracking-wide text-faint-foreground uppercase">
                      {dict.monitor.latency}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {filtered.map((event) => (
                    <EventRow
                      key={event.id}
                      event={event}
                      selected={event.id === selectedId}
                      onSelect={() => setSelectedId(event.id)}
                    />
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </Panel>
      </div>

      <div className="space-y-4">
        <Panel>
          <PanelHeader title={dict.monitor.counters} />
          <PanelBody className="grid grid-cols-2 gap-3">
            <Stat label={dict.overview.successfulReads} value={counters.successfulReads} mono tone="success" />
            <Stat label={dict.overview.failedReads} value={counters.failedReads} mono tone={counters.failedReads > 0 ? "danger" : "muted"} />
            <Stat label={dict.overview.retries} value={counters.retries} mono tone={counters.retries > 0 ? "warning" : "muted"} />
            <Stat label={dict.overview.protocolErrors} value={counters.protocolErrors} mono tone={counters.protocolErrors > 0 ? "warning" : "muted"} />
            <Stat label={dict.overview.reconnects} value={counters.reconnects} mono tone={counters.reconnects > 0 ? "warning" : "muted"} />
            <Stat label={dict.overview.sessions} value={counters.connections} mono tone="muted" />
          </PanelBody>
        </Panel>

        <Panel className="xl:sticky xl:top-[4.5rem]">
          <PanelHeader title={dict.monitor.detail} hint={selected ? undefined : dict.monitor.detailHint} />
          <PanelBody>
            {selected && selected.bytes.length > 0 ? (
              <FrameInspector bytes={selected.bytes} mode={checksumMode} compact />
            ) : selected ? (
              <SystemEventDetail event={selected} />
            ) : (
              <p className="text-xs text-faint-foreground">{dict.monitor.detailHint}</p>
            )}
          </PanelBody>
        </Panel>
      </div>
    </div>
  );
}

function EventRow({
  event,
  selected,
  onSelect,
}: {
  event: LogEvent;
  selected: boolean;
  onSelect: () => void;
}) {
  const dict = useDictionary();
  const format = useFormat();

  return (
    <tr
      onClick={onSelect}
      aria-selected={selected}
      className={cn(
        "cursor-pointer border-b border-border/50 transition-colors",
        selected ? "bg-primary/10" : "hover:bg-surface-raised/60",
      )}
    >
      <td className="w-[5.5rem] py-1 pl-3.5 align-top font-mono text-[0.6875rem] text-faint-foreground tabular">
        {format.clock(event.at)}
      </td>
      <td className="w-11 py-1 align-top">
        <span className={cn("font-mono text-[0.6875rem] font-semibold", DIRECTION_TEXT[event.direction])}>
          {dict.directions[event.direction === "sys" ? "system" : event.direction]}
        </span>
      </td>
      <td className="w-[7.5rem] py-1 align-top">
        <span className="truncate text-[0.6875rem] text-muted-foreground">{event.command}</span>
      </td>
      <td className="py-1 pr-3.5 align-top">
        {event.bytes.length > 0 ? (
          <span className="block truncate font-mono text-[0.6875rem] text-foreground tabular">
            {formatHex(event.bytes)}
          </span>
        ) : event.errorCode || (event.direction === "err" && event.detail) ? (
          <span className="text-[0.6875rem] text-destructive">
            {event.errorCode ? dict.events.errors[event.errorCode] : event.detail}
          </span>
        ) : event.note ? (
          <span className="text-[0.6875rem] text-signal-sys">
            {dict.events[event.note]}
            {event.noteArgs ? (
              <span className="ml-1 text-faint-foreground">
                {stateLabel(dict, event.noteArgs.from)} → {stateLabel(dict, event.noteArgs.to)}
              </span>
            ) : null}
          </span>
        ) : event.detail ? (
          <span className="block truncate text-[0.6875rem] text-signal-sys">{event.detail}</span>
        ) : null}
      </td>
      <td className="w-14 py-1 pr-3.5 text-right align-top font-mono text-[0.6875rem] text-faint-foreground tabular">
        {event.latencyMs !== undefined ? format.duration(event.latencyMs) : ""}
      </td>
    </tr>
  );
}

function SystemEventDetail({ event }: { event: LogEvent }) {
  const dict = useDictionary();
  const format = useFormat();

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <Badge tone={DIRECTION_TONE[event.direction]}>{event.command}</Badge>
        <span className="font-mono text-xs text-faint-foreground tabular">
          {format.clock(event.at)}
        </span>
      </div>
      <p className="text-sm break-words text-foreground">
        {event.errorCode
          ? dict.events.errors[event.errorCode]
          : event.note
            ? dict.events[event.note]
            : (event.detail ?? dict.common.none)}
      </p>
      {/* The source's own words, kept beside the translation rather than
          instead of it: a gateway message names the sentinel that produced it,
          which is what someone reading the Go log needs to match against. */}
      {event.detail && (event.errorCode || event.note) ? (
        <p className="font-mono text-[0.6875rem] break-words text-faint-foreground">
          {event.detail}
        </p>
      ) : null}
      {event.noteArgs ? (
        <p className="text-xs text-faint-foreground">
          {stateLabel(dict, event.noteArgs.from)} → {stateLabel(dict, event.noteArgs.to)}
        </p>
      ) : null}
    </div>
  );
}

/**
 * Sequence diagram over the last few cycles.
 *
 * Two lanes, gateway on top and device below, with every frame drawn at its
 * real offset in milliseconds. This is the view that makes a retry legible:
 * three requests and one answer in the space one exchange should have taken is
 * a shape, not a number.
 */
function SequenceDiagram({ events }: { events: LogEvent[] }) {
  const dict = useDictionary();
  const format = useFormat();

  const frames = useMemo(() => events.filter((event) => event.direction !== "sys").slice(-40), [events]);
  if (frames.length === 0) {
    return <p className="text-xs text-faint-foreground">{dict.monitor.emptyHint}</p>;
  }

  const start = frames[0]!.at;
  const end = frames[frames.length - 1]!.at;
  const span = Math.max(1, end - start);

  return (
    <div className="space-y-2">
      <div className="relative h-24 rounded border border-border bg-surface-sunken">
        <div className="absolute inset-x-0 top-7 h-px bg-border" />
        <div className="absolute inset-x-0 bottom-7 h-px bg-border" />
        <span className="absolute top-1.5 left-2 text-[0.625rem] tracking-wide text-faint-foreground uppercase">
          gateway
        </span>
        <span className="absolute bottom-1.5 left-2 text-[0.625rem] tracking-wide text-faint-foreground uppercase">
          device
        </span>

        {frames.map((event) => {
          const left = ((event.at - start) / span) * 96 + 2;
          const isFromGateway = event.source === "gateway";
          return (
            <div
              key={event.id}
              className="absolute"
              style={{ left: `${left}%`, top: "1.75rem", height: "3.5rem" }}
              title={`${event.command} · ${format.clock(event.at)}`}
            >
              <div
                className={cn(
                  "h-full w-px",
                  DIRECTION_BG[event.direction],
                  event.direction === "err" ? "opacity-90" : "opacity-60",
                )}
              />
              <div
                className={cn(
                  "absolute size-1.5 -translate-x-[0.1875rem] rounded-full",
                  DIRECTION_BG[event.direction],
                  isFromGateway ? "top-0" : "bottom-0",
                )}
              />
            </div>
          );
        })}
      </div>
      <div className="flex flex-wrap gap-x-4 gap-y-1">
        {(["tx", "rx", "err"] as const).map((direction) => (
          <span key={direction} className="flex items-center gap-1.5 text-[0.6875rem] text-faint-foreground">
            <span className={cn("size-2 rounded-[2px]", DIRECTION_BG[direction])} />
            {dict.directions[`${direction}Full` as "txFull"]}
          </span>
        ))}
        <span className="ml-auto font-mono text-[0.6875rem] text-faint-foreground tabular">
          {format.duration(span)}
        </span>
      </div>
    </div>
  );
}

/**
 * Names a state carried in an event's arguments.
 *
 * The engine records transitions as raw state names because it has no
 * dictionary; translating them here keeps the log readable in both languages
 * without teaching the store about locales.
 */
function stateLabel(dict: ReturnType<typeof useDictionary>, state: string): string {
  return state in dict.states ? dict.states[state as keyof typeof dict.states] : state;
}
