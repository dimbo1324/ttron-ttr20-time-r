"use client";

import {
  AlertTriangle,
  ArrowLeftRight,
  BookOpen,
  Calculator,
  Binary,
  GraduationCap,
  Layers,
  ListTree,
  Waves,
  Zap,
} from "lucide-react";
import { useMemo, useState, type ReactNode } from "react";

import { useDictionary } from "@/components/locale-provider";
import { FrameLayoutBar } from "@/components/protocol/frame-view";
import { Badge } from "@/components/ui/badge";
import { Field, SegmentedControl, TextArea } from "@/components/ui/controls";
import { Panel, PanelBody, PanelHeader } from "@/components/ui/panel";
import {
  buildReadTimeRequest,
  CHECKSUM_MODES,
  COMMANDS,
  crc16,
  decodeFrame,
  encodeFrame,
  formatHex,
  parseHex,
  sum8,
  toHex,
  type ChecksumMode,
} from "@/lib/ft12";
import { DIRECTION_TEXT } from "@/lib/status";
import { cn } from "@/lib/utils";

/**
 * The reference.
 *
 * This is the page that has to work for someone who has never seen a protocol
 * frame, so it is written as prose in reading order — what the equipment is,
 * why anyone asks it for the time, what a byte is, and only then how the frame
 * is laid out. A specification that begins with the frame diagram teaches
 * nobody who did not already know.
 *
 * Two things here are live rather than described, because a claim a reader can
 * check is worth more than one they must accept: the frame layout is rendered
 * by the same decoder the analyzer uses, and the checksum calculator recomputes
 * with the same functions the wire does.
 */

const SECTION_IDS = [
  "basics",
  "exchange",
  "numbers",
  "frame",
  "checksums",
  "commands",
  "stream",
  "faults",
  "zone",
  "glossary",
] as const;

type SectionId = (typeof SECTION_IDS)[number];

export function ProtocolReference() {
  const dict = useDictionary();
  const [mode, setMode] = useState<ChecksumMode>("sum");
  const [payloadText, setPayloadText] = useState("00 01 01");

  const payload = useMemo(() => parseHex(payloadText).bytes, [payloadText]);
  const sample = useMemo(
    () => decodeFrame(encodeFrame(0x00, 0x01, buildReadTimeRequest(), mode), mode),
    [mode],
  );
  const crcValue = crc16(payload);

  const titles: Record<SectionId, string> = {
    basics: dict.reference.basics.title,
    exchange: dict.reference.exchange.title,
    numbers: dict.reference.numbers.title,
    frame: dict.reference.frame.title,
    checksums: dict.reference.checksums.title,
    commands: dict.reference.commands.title,
    stream: dict.reference.stream.title,
    faults: dict.reference.faults.title,
    zone: dict.reference.zone.title,
    glossary: dict.reference.glossary.title,
  };

  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_18rem]">
      <article className="min-w-0 space-y-4">
        <Section id="basics" icon={<GraduationCap className="size-4" />} title={titles.basics}>
          <p className="text-[0.9375rem] leading-relaxed text-foreground">
            {dict.reference.basics.lead}
          </p>
          <Prose paragraphs={dict.reference.basics.body} />
        </Section>

        <Section id="exchange" icon={<ArrowLeftRight className="size-4" />} title={titles.exchange}>
          <Prose paragraphs={dict.reference.exchange.body} />
          <dl className="mt-1 grid gap-1.5 sm:grid-cols-2">
            {dict.reference.exchange.lanes.map((lane) => (
              <div
                key={lane.label}
                className="flex items-baseline gap-2 rounded border border-border bg-surface-sunken px-2.5 py-1.5"
              >
                <dt
                  className={cn(
                    "shrink-0 font-mono text-xs font-semibold",
                    DIRECTION_TEXT[lane.label.toLowerCase() as keyof typeof DIRECTION_TEXT] ??
                      "text-foreground",
                  )}
                >
                  {lane.label}
                </dt>
                <dd className="min-w-0 text-xs text-muted-foreground">{lane.meaning}</dd>
              </div>
            ))}
          </dl>
        </Section>

        <Section id="numbers" icon={<Binary className="size-4" />} title={titles.numbers}>
          <Prose paragraphs={dict.reference.numbers.body} />
          <div className="grid gap-1.5 sm:grid-cols-4">
            {[
              { hex: "0x00", dec: "0" },
              { hex: "0x0F", dec: "15" },
              { hex: "0x68", dec: "104" },
              { hex: "0xFF", dec: "255" },
            ].map((row) => (
              <div
                key={row.hex}
                className="rounded border border-border bg-surface-sunken px-2.5 py-2 text-center"
              >
                <div className="font-mono text-sm text-primary tabular">{row.hex}</div>
                <div className="font-mono text-xs text-faint-foreground tabular">= {row.dec}</div>
              </div>
            ))}
          </div>
        </Section>

        <Section id="frame" icon={<Layers className="size-4" />} title={titles.frame}>
          <p className="text-[0.9375rem] leading-relaxed text-foreground">
            {dict.reference.frame.lead}
          </p>
          <code className="block overflow-x-auto rounded border border-border bg-surface-sunken px-3 py-2.5 font-mono text-xs text-primary">
            0x68 | LEN | 0x68 | CONTROL | ADDRESS | DATA… | CHECKSUM | 0x16
          </code>
          <Prose paragraphs={dict.reference.frame.body} />
          <FrameLayoutBar decoded={sample} />

          <div className="pt-1">
            <h3 className="text-sm font-semibold text-foreground">
              {dict.reference.frame.walkthrough.title}
            </h3>
            <p className="mt-0.5 font-mono text-xs text-faint-foreground">
              {dict.reference.frame.walkthrough.hint}
            </p>

            <div className="mt-2 overflow-x-auto">
              <table className="w-full border-collapse text-left">
                <thead>
                  <tr className="border-b border-border">
                    <Th>{dict.reference.frame.walkthrough.columns.byte}</Th>
                    <Th>{dict.reference.frame.walkthrough.columns.name}</Th>
                    <Th>{dict.reference.frame.walkthrough.columns.meaning}</Th>
                  </tr>
                </thead>
                <tbody>
                  {dict.reference.frame.walkthrough.rows.map((row, index) => (
                    <tr key={`${row.byte}-${index}`} className="border-b border-border/50 align-top">
                      <td className="w-12 py-1.5 pr-3 font-mono text-sm font-semibold text-primary tabular">
                        {row.byte}
                      </td>
                      <td className="w-44 py-1.5 pr-3 text-xs font-medium text-foreground">
                        {row.name}
                      </td>
                      <td className="py-1.5 text-xs leading-snug text-muted-foreground">
                        {row.meaning}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </Section>

        <Section
          id="checksums"
          icon={<Calculator className="size-4" />}
          title={titles.checksums}
          actions={
            <SegmentedControl
              value={mode}
              onChange={setMode}
              options={CHECKSUM_MODES.map((item) => ({ value: item, label: item }))}
            />
          }
        >
          <Prose paragraphs={dict.reference.checksums.body} />

          <div className="grid gap-2 sm:grid-cols-2">
            {dict.reference.checksums.modes.map((item) => (
              <div key={item.name} className="rounded border border-border bg-surface-sunken p-2.5">
                <Badge tone="primary" mono>
                  {item.name}
                </Badge>
                <p className="mt-1.5 text-xs leading-snug text-muted-foreground">{item.body}</p>
              </div>
            ))}
          </div>

          <div className="rounded border border-border p-2.5">
            <h3 className="text-sm font-semibold text-foreground">
              {dict.reference.checksums.calculator}
            </h3>
            <div className="mt-2">
              <Field
                label={dict.reference.checksums.calculatorInput}
                hint={dict.reference.checksums.calculatorHint}
              >
                <TextArea
                  value={payloadText}
                  onChange={(event) => setPayloadText(event.target.value)}
                  rows={2}
                  spellCheck={false}
                  className="font-mono text-xs tabular"
                />
              </Field>
            </div>

            <div className="mt-3 grid gap-2 sm:grid-cols-2">
              <Readout label="sum8" value={toHex(sum8(payload))} />
              <Readout
                label="crc16 · little-endian"
                value={formatHex([crcValue & 0xff, (crcValue >> 8) & 0xff])}
              />
            </div>
          </div>
        </Section>

        <Section id="commands" icon={<ListTree className="size-4" />} title={titles.commands}>
          <Prose paragraphs={dict.reference.commands.body} />
          {COMMANDS.map((command) => (
            <div key={command.id} className="rounded border border-border bg-surface-sunken p-2.5">
              <div className="flex flex-wrap items-center gap-2">
                <Badge tone="primary" mono>
                  0x{toHex(command.id)}
                </Badge>
                <span className="font-mono text-xs text-foreground">{command.name}</span>
                <span className="text-xs text-muted-foreground">
                  {dict.protocol.commandInfo[command.descriptionKey]}
                </span>
              </div>
              <div className="mt-2 grid gap-1.5 sm:grid-cols-2">
                <Readout
                  label={dict.reference.commands.columns.request}
                  value={`DATA = 0x${toHex(command.id)}`}
                  mono
                />
                <Readout
                  label={dict.reference.commands.columns.response}
                  value={
                    command.id === 0x01
                      ? 'DATA = 0x01 + "YYYY-MM-DD HH:MM:SS"'
                      : 'DATA = 0x02 + "MODEL|SERIAL|FIRMWARE"'
                  }
                  mono
                />
              </div>
            </div>
          ))}
        </Section>

        <Section id="stream" icon={<Waves className="size-4" />} title={titles.stream}>
          <Prose paragraphs={dict.reference.stream.body} />
        </Section>

        <Section id="faults" icon={<Zap className="size-4" />} title={titles.faults}>
          <Prose paragraphs={dict.reference.faults.body} />
          <dl className="space-y-1">
            {dict.reference.faults.rows.map((row) => (
              <div
                key={row.name}
                className="border-b border-border/60 py-1.5 last:border-0 sm:flex sm:gap-3"
              >
                <dt className="shrink-0 text-xs font-medium text-foreground sm:w-52">{row.name}</dt>
                <dd className="text-xs leading-snug text-muted-foreground">{row.meaning}</dd>
              </div>
            ))}
          </dl>
        </Section>

        <Section
          id="zone"
          icon={<AlertTriangle className="size-4 text-warning" />}
          title={titles.zone}
          className="border-warning/30"
        >
          <Prose paragraphs={dict.reference.zone.body} />
        </Section>

        <Section id="glossary" icon={<BookOpen className="size-4" />} title={titles.glossary}>
          <p className="text-xs text-faint-foreground">{dict.reference.glossary.hint}</p>
          <dl className="grid gap-x-4 sm:grid-cols-2">
            {dict.reference.glossary.terms.map((entry) => (
              <div key={entry.term} className="border-b border-border/60 py-2 last:border-0">
                <dt className="text-xs font-semibold text-foreground">{entry.term}</dt>
                <dd className="mt-0.5 text-xs leading-snug text-muted-foreground">
                  {entry.definition}
                </dd>
              </div>
            ))}
          </dl>
        </Section>
      </article>

      <aside className="space-y-4 xl:sticky xl:top-[4.5rem] xl:self-start">
        <Panel>
          <PanelHeader title={dict.reference.title} hint={dict.reference.subtitle} />
          <PanelBody>
            <nav aria-label={dict.reference.title}>
              <ol className="space-y-0.5">
                {SECTION_IDS.map((id, index) => (
                  <li key={id}>
                    <a
                      href={`#${id}`}
                      className="flex items-baseline gap-2 rounded px-1.5 py-1 text-xs text-muted-foreground transition-colors hover:bg-surface-raised hover:text-foreground"
                    >
                      <span className="font-mono text-[0.625rem] text-faint-foreground tabular">
                        {String(index + 1).padStart(2, "0")}
                      </span>
                      <span className="min-w-0 truncate">{titles[id]}</span>
                    </a>
                  </li>
                ))}
              </ol>
            </nav>
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader title={dict.reference.errorsTitle} hint={dict.reference.errorsHint} />
          <PanelBody className="space-y-1">
            {Object.entries(dict.protocol.errors).map(([code, label]) => (
              <div
                key={code}
                className="flex items-baseline justify-between gap-3 border-b border-border/60 py-1.5 last:border-0"
              >
                <code className="shrink-0 font-mono text-[0.6875rem] text-destructive">{code}</code>
                <span className="min-w-0 text-right text-xs text-muted-foreground">{label}</span>
              </div>
            ))}
          </PanelBody>
        </Panel>
      </aside>
    </div>
  );
}

function Section({
  id,
  icon,
  title,
  actions,
  className,
  children,
}: {
  id: SectionId;
  icon: ReactNode;
  title: string;
  actions?: ReactNode;
  className?: string;
  children: ReactNode;
}) {
  return (
    // `scroll-mt` keeps the heading clear of the sticky status strip when the
    // contents rail jumps to it.
    <Panel id={id} className={cn("scroll-mt-20", className)}>
      <PanelHeader icon={icon} title={title} actions={actions} />
      <PanelBody className="space-y-3">{children}</PanelBody>
    </Panel>
  );
}

/** Body copy, at a measure that stays readable on a wide instrument screen. */
function Prose({ paragraphs }: { paragraphs: readonly string[] }) {
  return (
    <div className="space-y-2">
      {paragraphs.map((paragraph, index) => (
        <p key={index} className="max-w-[68ch] text-[0.8125rem] leading-relaxed text-muted-foreground">
          {paragraph}
        </p>
      ))}
    </div>
  );
}

function Readout({ label, value, mono = true }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="rounded border border-border bg-surface-sunken px-2.5 py-2">
      <div className="text-[0.625rem] tracking-wide text-faint-foreground uppercase">{label}</div>
      <div className={cn("text-foreground", mono ? "font-mono text-xs tabular" : "text-sm")}>
        {value}
      </div>
    </div>
  );
}

function Th({ children }: { children: ReactNode }) {
  return (
    <th
      scope="col"
      className="py-1 pr-3 text-[0.625rem] font-medium tracking-wide text-faint-foreground uppercase"
    >
      {children}
    </th>
  );
}
