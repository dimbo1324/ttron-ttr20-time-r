"use client";

import { AlertTriangle, CheckCircle2 } from "lucide-react";
import { useMemo, useState } from "react";

import { useDictionary } from "@/components/locale-provider";
import { Badge } from "@/components/ui/badge";
import {
  asciiGlyph,
  decodeFrame,
  formatBinary,
  formatHex,
  parsePayload,
  RESPONSE_BIT,
  toHex,
  type ChecksumMode,
  type DecodedFrame,
  type FieldKind,
} from "@/lib/ft12";
import { cn } from "@/lib/utils";

/**
 * The frame views.
 *
 * These are the components the whole bench exists to show, so they are built
 * around one idea: a frame is not a string of hex, it is a sequence of fields,
 * and every view here colours the same byte the same way. A reader who learns
 * "amber is the checksum" on the byte map reads it instantly on the layout bar
 * and in the log.
 */

const FIELD_STYLE: Record<FieldKind, { text: string; bg: string; border: string; bar: string }> = {
  start: {
    text: "text-field-frame",
    bg: "bg-field-frame/10",
    border: "border-field-frame/30",
    bar: "bg-field-frame/45",
  },
  length: {
    text: "text-field-length",
    bg: "bg-field-length/12",
    border: "border-field-length/35",
    bar: "bg-field-length/60",
  },
  startRepeat: {
    text: "text-field-frame",
    bg: "bg-field-frame/10",
    border: "border-field-frame/30",
    bar: "bg-field-frame/45",
  },
  control: {
    text: "text-field-control",
    bg: "bg-field-control/12",
    border: "border-field-control/35",
    bar: "bg-field-control/60",
  },
  address: {
    text: "text-field-address",
    bg: "bg-field-address/12",
    border: "border-field-address/35",
    bar: "bg-field-address/60",
  },
  data: {
    text: "text-field-data",
    bg: "bg-field-data/12",
    border: "border-field-data/35",
    bar: "bg-field-data/60",
  },
  checksum: {
    text: "text-field-checksum",
    bg: "bg-field-checksum/12",
    border: "border-field-checksum/35",
    bar: "bg-field-checksum/60",
  },
  end: {
    text: "text-field-end",
    bg: "bg-field-end/10",
    border: "border-field-end/30",
    bar: "bg-field-end/45",
  },
  trailing: {
    text: "text-destructive",
    bg: "bg-destructive/10",
    border: "border-destructive/30",
    bar: "bg-destructive/45",
  },
};

function useFieldLabels() {
  const dict = useDictionary();
  return dict.protocol.fields as Record<FieldKind, string>;
}

/**
 * Proportional map of the whole frame.
 *
 * Widths are proportional to byte count, which makes the one structural fact
 * that trips people up self-evident: DATA dominates the frame, and the header
 * and checksum are a fixed, small tax around it.
 */
export function FrameLayoutBar({ decoded }: { decoded: DecodedFrame }) {
  const labels = useFieldLabels();
  const total = decoded.fields.reduce((sum, field) => sum + field.bytes.length, 0);
  if (total === 0) return null;

  return (
    <div className="space-y-1.5">
      <div className="flex h-7 w-full overflow-hidden rounded border border-border">
        {decoded.fields.map((field) => (
          <div
            key={`${field.kind}-${field.offset}`}
            style={{ width: `${(field.bytes.length / total) * 100}%` }}
            className={cn(
              "flex items-center justify-center border-r border-background/40 last:border-r-0",
              FIELD_STYLE[field.kind].bar,
            )}
            title={`${labels[field.kind]} · ${field.bytes.length}`}
          >
            {field.bytes.length > 1 ? (
              <span className="truncate px-1 font-mono text-[0.625rem] font-medium text-background">
                {field.bytes.length}
              </span>
            ) : null}
          </div>
        ))}
      </div>
      <div className="flex flex-wrap gap-x-3 gap-y-1">
        {decoded.fields.map((field) => (
          <span
            key={`legend-${field.kind}-${field.offset}`}
            className="flex items-center gap-1.5 text-[0.6875rem] text-faint-foreground"
          >
            <span className={cn("size-2 rounded-[2px]", FIELD_STYLE[field.kind].bar)} />
            {labels[field.kind]}
          </span>
        ))}
      </div>
    </div>
  );
}

interface ByteCell {
  offset: number;
  value: number;
  kind: FieldKind;
}

/**
 * Byte-level view.
 *
 * Each cell carries hex, and hovering reveals the binary and ASCII of that
 * byte in the inspector strip below — the two representations a reader needs
 * exactly when a field looks wrong (a flipped bit, a stray printable
 * character) and never wants filling the grid the rest of the time.
 */
export function ByteGrid({ decoded }: { decoded: DecodedFrame }) {
  const dict = useDictionary();
  const labels = useFieldLabels();
  const [hovered, setHovered] = useState<number | null>(null);

  const cells = useMemo<ByteCell[]>(
    () =>
      decoded.fields.flatMap((field) =>
        field.bytes.map((value, index) => ({
          offset: field.offset + index,
          value,
          kind: field.kind,
        })),
      ),
    [decoded.fields],
  );

  const badOffsets = useMemo(
    () => new Set(decoded.issues.map((issue) => issue.offset).filter((offset) => offset !== undefined)),
    [decoded.issues],
  );

  const active = hovered !== null ? cells.find((cell) => cell.offset === hovered) : undefined;

  if (cells.length === 0) return null;

  return (
    <div className="space-y-2.5">
      <div className="flex flex-wrap gap-1">
        {cells.map((cell) => {
          const style = FIELD_STYLE[cell.kind];
          const bad = badOffsets.has(cell.offset);
          return (
            <button
              key={cell.offset}
              type="button"
              onMouseEnter={() => setHovered(cell.offset)}
              onMouseLeave={() => setHovered((current) => (current === cell.offset ? null : current))}
              onFocus={() => setHovered(cell.offset)}
              onBlur={() => setHovered((current) => (current === cell.offset ? null : current))}
              className={cn(
                "group relative flex w-11 flex-col items-center rounded border py-1 outline-none transition-colors duration-150",
                style.bg,
                bad ? "border-destructive ring-1 ring-destructive/40" : style.border,
                "hover:border-current focus-visible:ring-[2px] focus-visible:ring-ring/50",
              )}
            >
              <span className="font-mono text-[0.5625rem] text-faint-foreground tabular">
                {cell.offset.toString().padStart(2, "0")}
              </span>
              <span className={cn("font-mono text-[0.8125rem] font-semibold tabular", style.text)}>
                {toHex(cell.value)}
              </span>
            </button>
          );
        })}
      </div>

      <div className="flex min-h-[2.25rem] flex-wrap items-center gap-x-5 gap-y-1 rounded border border-border bg-surface-sunken px-2.5 py-1.5">
        {active ? (
          <>
            <InspectorItem label={dict.protocol.field} value={labels[active.kind]} />
            <InspectorItem label={dict.protocol.offset} value={String(active.offset)} mono />
            <InspectorItem label={dict.protocol.hex} value={`0x${toHex(active.value)}`} mono />
            <InspectorItem label={dict.protocol.bin} value={formatBinary(active.value)} mono />
            <InspectorItem label={dict.protocol.ascii} value={asciiGlyph(active.value)} mono />
          </>
        ) : (
          <span className="text-xs text-faint-foreground">{dict.protocol.byteMap}</span>
        )}
      </div>
    </div>
  );
}

function InspectorItem({
  label,
  value,
  mono = false,
}: {
  label: string;
  value: string;
  mono?: boolean;
}) {
  return (
    <span className="flex items-baseline gap-1.5">
      <span className="text-[0.625rem] tracking-wide text-faint-foreground uppercase">{label}</span>
      <span className={cn("text-xs text-foreground", mono && "font-mono tabular")}>{value}</span>
    </span>
  );
}

/** Verdict strip: valid, or the exact reason it is not. */
export function FrameVerdict({ decoded }: { decoded: DecodedFrame }) {
  const dict = useDictionary();
  const errors = dict.protocol.errors as Record<string, string>;

  if (decoded.issues.length === 0) {
    return (
      <div className="flex items-center gap-2 rounded border border-success/30 bg-success/10 px-2.5 py-1.5 text-xs text-success">
        <CheckCircle2 className="size-3.5 shrink-0" />
        {dict.protocol.valid}
      </div>
    );
  }

  return (
    <div className="space-y-1.5">
      {decoded.issues.map((issue, index) => (
        <div
          key={`${issue.code}-${index}`}
          className="flex items-start gap-2 rounded border border-destructive/30 bg-destructive/10 px-2.5 py-1.5 text-xs text-destructive"
        >
          <AlertTriangle className="mt-px size-3.5 shrink-0" />
          <div className="min-w-0">
            <span>{errors[issue.code] ?? issue.code}</span>
            {issue.expected && issue.actual ? (
              <span className="ml-1.5 font-mono text-[0.6875rem] text-destructive/80 tabular">
                ({dict.protocol.expected} {formatHex(issue.expected)} · {dict.protocol.actual}{" "}
                {formatHex(issue.actual)})
              </span>
            ) : null}
            {issue.offset !== undefined ? (
              <span className="ml-1.5 font-mono text-[0.6875rem] text-destructive/70">
                @{issue.offset}
              </span>
            ) : null}
          </div>
        </div>
      ))}
    </div>
  );
}

/** What the DATA field means, once the frame itself is understood. */
export function CommandDecode({ decoded }: { decoded: DecodedFrame }) {
  const dict = useDictionary();
  const errors = dict.protocol.errors as Record<string, string>;

  const control = decoded.fields.find((field) => field.kind === "control")?.bytes[0];
  const data = decoded.fields.find((field) => field.kind === "data")?.bytes ?? [];
  if (control === undefined) return null;

  const isResponse = (control & RESPONSE_BIT) !== 0;
  const parsed = parsePayload(data, isResponse);

  return (
    <div className="space-y-2">
      <div className="flex flex-wrap items-center gap-2">
        <Badge tone={isResponse ? "rx" : "tx"} mono>
          {isResponse ? dict.common.response : dict.common.request}
        </Badge>
        <Badge tone="primary" mono>
          {parsed.commandName}
        </Badge>
        {parsed.commandId !== undefined ? (
          <span className="font-mono text-xs text-faint-foreground tabular">
            0x{toHex(parsed.commandId)}
          </span>
        ) : null}
      </div>

      {parsed.error ? (
        <p className="text-xs text-destructive">{errors[parsed.error] ?? parsed.error}</p>
      ) : null}

      {parsed.fields ? (
        <div className="grid gap-1.5 sm:grid-cols-3">
          {parsed.fields.map((field) => (
            <div key={field.label} className="rounded border border-border bg-surface-sunken px-2 py-1.5">
              <div className="text-[0.625rem] tracking-wide text-faint-foreground uppercase">
                {field.label}
              </div>
              <div className="truncate font-mono text-xs text-foreground">{field.value}</div>
            </div>
          ))}
        </div>
      ) : parsed.value ? (
        <div className="rounded border border-border bg-surface-sunken px-2.5 py-2 font-mono text-sm text-foreground tabular">
          {parsed.value}
        </div>
      ) : null}
    </div>
  );
}

/** Everything above, stacked — used by the analyzer and the monitor detail. */
export function FrameInspector({
  bytes,
  mode,
  compact = false,
}: {
  bytes: number[];
  mode: ChecksumMode;
  compact?: boolean;
}) {
  const dict = useDictionary();
  const decoded = useMemo(() => decodeFrame(bytes, mode), [bytes, mode]);

  if (bytes.length === 0) {
    return <p className="text-xs text-faint-foreground">{dict.protocol.empty}</p>;
  }

  return (
    <div className="space-y-3.5">
      <FrameVerdict decoded={decoded} />
      {!compact ? <FrameLayoutBar decoded={decoded} /> : null}
      <ByteGrid decoded={decoded} />
      <CommandDecode decoded={decoded} />
      {decoded.expectedChecksum.length > 0 ? (
        <div className="flex flex-wrap items-center gap-2 border-t border-border pt-2.5 text-xs">
          <span className="text-faint-foreground">{dict.protocol.computed}</span>
          <span className="font-mono text-field-checksum tabular">
            {formatHex(decoded.expectedChecksum)}
          </span>
          <span className="text-faint-foreground">· {dict.protocol.payloadSpan}</span>
        </div>
      ) : null}
    </div>
  );
}

export { FIELD_STYLE };
