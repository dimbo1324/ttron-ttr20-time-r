"use client";

import { BookOpen, Calculator, Globe2, ListTree } from "lucide-react";
import { useMemo, useState } from "react";

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

/**
 * Protocol reference.
 *
 * The teaching page. It states the format once, then gives the reader the two
 * things a specification cannot: a live checksum calculator, so the arithmetic
 * stops being a claim, and the time-zone note, which is the one part of this
 * protocol that has already produced a real defect and would produce it again
 * in any second implementation.
 */
export function ProtocolReference() {
  const dict = useDictionary();
  const [mode, setMode] = useState<ChecksumMode>("sum");
  const [payloadText, setPayloadText] = useState("00 01 01");

  const payload = useMemo(() => parseHex(payloadText).bytes, [payloadText]);
  const sample = useMemo(() => decodeFrame(encodeFrame(0x00, 0x01, buildReadTimeRequest(), mode), mode), [mode]);

  const sumValue = sum8(payload);
  const crcValue = crc16(payload);

  const errorEntries = Object.entries(dict.protocol.errors) as [string, string][];

  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_22rem]">
      <div className="space-y-4">
        <Panel>
          <PanelHeader
            icon={<BookOpen className="size-4" />}
            title={dict.reference.frameFormat}
            hint={dict.reference.subtitle}
          />
          <PanelBody className="space-y-3.5">
            <code className="block overflow-x-auto rounded border border-border bg-surface-sunken px-3 py-2.5 font-mono text-xs text-primary">
              0x68 | LEN | 0x68 | CONTROL | ADDRESS | DATA… | CHECKSUM | 0x16
            </code>
            <p className="text-[0.8125rem] leading-relaxed text-muted-foreground">
              {dict.reference.frameFormatBody}
            </p>
            <div className="border-t border-border pt-3">
              <FrameLayoutBar decoded={sample} />
            </div>
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader
            icon={<Calculator className="size-4" />}
            title={dict.reference.checksums}
            hint={dict.reference.checksumSpan}
            actions={
              <SegmentedControl
                value={mode}
                onChange={setMode}
                options={CHECKSUM_MODES.map((item) => ({ value: item, label: item }))}
              />
            }
          />
          <PanelBody className="space-y-3">
            <div className="grid gap-2 sm:grid-cols-2">
              <div className="rounded border border-border bg-surface-sunken p-2.5">
                <Badge tone="primary" mono>
                  sum
                </Badge>
                <p className="mt-1.5 text-xs leading-snug text-muted-foreground">
                  {dict.reference.checksumSum}
                </p>
              </div>
              <div className="rounded border border-border bg-surface-sunken p-2.5">
                <Badge tone="primary" mono>
                  crc16
                </Badge>
                <p className="mt-1.5 text-xs leading-snug text-muted-foreground">
                  {dict.reference.checksumCrc}
                </p>
              </div>
            </div>

            <Field label={dict.reference.calculatorInput} hint={dict.reference.calculatorHint}>
              <TextArea
                value={payloadText}
                onChange={(event) => setPayloadText(event.target.value)}
                rows={2}
                spellCheck={false}
                className="font-mono text-xs tabular"
              />
            </Field>

            <div className="grid gap-2 sm:grid-cols-2">
              <div className="rounded border border-border bg-surface-sunken px-3 py-2">
                <div className="text-[0.625rem] tracking-wide text-faint-foreground uppercase">
                  sum8
                </div>
                <div className="font-mono text-lg text-field-checksum tabular">
                  {toHex(sumValue)}
                </div>
              </div>
              <div className="rounded border border-border bg-surface-sunken px-3 py-2">
                <div className="text-[0.625rem] tracking-wide text-faint-foreground uppercase">
                  crc16 · little-endian
                </div>
                <div className="font-mono text-lg text-field-checksum tabular">
                  {formatHex([crcValue & 0xff, (crcValue >> 8) & 0xff])}
                </div>
              </div>
            </div>
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader icon={<ListTree className="size-4" />} title={dict.reference.commands} />
          <PanelBody className="space-y-2">
            {COMMANDS.map((command) => (
              <div key={command.id} className="rounded border border-border bg-surface-sunken p-2.5">
                <div className="flex flex-wrap items-center gap-2">
                  <Badge tone="primary" mono>
                    0x{toHex(command.id)}
                  </Badge>
                  <span className="font-mono text-xs text-foreground">{command.name}</span>
                  <span className="text-xs text-faint-foreground">
                    {dict.protocol.commandInfo[command.descriptionKey]}
                  </span>
                </div>
                <div className="mt-2 grid gap-1.5 sm:grid-cols-2">
                  <div>
                    <div className="text-[0.625rem] tracking-wide text-faint-foreground uppercase">
                      {dict.reference.requestFormat}
                    </div>
                    <code className="font-mono text-xs text-muted-foreground">
                      DATA = 0x{toHex(command.id)}
                    </code>
                  </div>
                  <div>
                    <div className="text-[0.625rem] tracking-wide text-faint-foreground uppercase">
                      {dict.reference.responseFormat}
                    </div>
                    <code className="font-mono text-xs text-muted-foreground">
                      {command.id === 0x01
                        ? 'DATA = 0x01 + "YYYY-MM-DD HH:MM:SS"'
                        : 'DATA = 0x02 + "MODEL|SERIAL|FIRMWARE"'}
                    </code>
                  </div>
                </div>
              </div>
            ))}
          </PanelBody>
        </Panel>
      </div>

      <div className="space-y-4">
        <Panel className="border-warning/30">
          <PanelHeader
            icon={<Globe2 className="size-4 text-warning" />}
            title={dict.reference.zoneTitle}
          />
          <PanelBody>
            <p className="text-[0.8125rem] leading-relaxed text-muted-foreground">
              {dict.reference.zoneBody}
            </p>
          </PanelBody>
        </Panel>

        <Panel className="xl:sticky xl:top-[4.5rem]">
          <PanelHeader title={dict.reference.errorsTitle} hint={dict.reference.errorsHint} />
          <PanelBody className="space-y-1">
            {errorEntries.map(([code, label]) => (
              <div
                key={code}
                className="flex items-baseline justify-between gap-3 border-b border-border/60 py-1.5 last:border-0"
              >
                <code className="shrink-0 font-mono text-[0.6875rem] text-destructive">{code}</code>
                <span className="min-w-0 truncate text-right text-xs text-muted-foreground">
                  {label}
                </span>
              </div>
            ))}
          </PanelBody>
        </Panel>
      </div>
    </div>
  );
}
