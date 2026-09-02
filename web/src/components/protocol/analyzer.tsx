"use client";

import { Binary, Layers, Wand2 } from "lucide-react";
import { useMemo, useState } from "react";

import { useDictionary } from "@/components/locale-provider";
import { FrameInspector } from "@/components/protocol/frame-view";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Field, SegmentedControl, Select, TextArea } from "@/components/ui/controls";
import { Panel, PanelBody, PanelHeader, Stat } from "@/components/ui/panel";
import {
  buildReadIdentityRequest,
  buildReadIdentityResponse,
  buildReadTimeRequest,
  buildReadTimeResponse,
  CHECKSUM_MODES,
  COMMANDS,
  encodeFrame,
  formatHex,
  parseHex,
  RESPONSE_BIT,
  scanFrames,
  toHex,
  type ChecksumMode,
} from "@/lib/ft12";
import { cn } from "@/lib/utils";

/**
 * Frame analyzer.
 *
 * Three things share one hex buffer: a builder that writes into it, a decoder
 * that explains it, and a stream scanner that pulls frames out of a longer
 * capture. Keeping them on one page (rather than as three tools) is what makes
 * the loop work — build a frame, break a byte, watch the checksum verdict flip
 * — which is the whole lesson the bench is trying to deliver.
 */
export function FrameAnalyzer() {
  const dict = useDictionary();
  const [mode, setMode] = useState<ChecksumMode>("sum");
  const [text, setText] = useState(() => formatHex(encodeFrame(0x00, 0x01, buildReadTimeRequest(), "sum")));

  const [control, setControl] = useState(0x00);
  const [address, setAddress] = useState(0x01);
  const [commandId, setCommandId] = useState<number>(COMMANDS[0]!.id);
  const [isResponse, setIsResponse] = useState(false);

  const parsed = useMemo(() => parseHex(text), [text]);
  const scan = useMemo(() => scanFrames(parsed.bytes, mode), [parsed.bytes, mode]);

  const builderBytes = useMemo(() => {
    const effectiveControl = isResponse ? control | RESPONSE_BIT : control & ~RESPONSE_BIT;
    const data =
      commandId === 0x02
        ? isResponse
          ? buildReadIdentityResponse("TTR20", "SN-0000042", "1.2.3")
          : buildReadIdentityRequest()
        : isResponse
          ? buildReadTimeResponse(new Date())
          : buildReadTimeRequest();
    return encodeFrame(effectiveControl, address, data, mode);
  }, [address, commandId, control, isResponse, mode]);

  const samples = useMemo(
    () => [
      {
        label: "read-time · request",
        bytes: encodeFrame(0x00, 0x01, buildReadTimeRequest(), mode),
      },
      {
        label: "read-time · response",
        bytes: encodeFrame(RESPONSE_BIT, 0x01, buildReadTimeResponse(new Date()), mode),
      },
      {
        label: "read-identity · response",
        bytes: encodeFrame(
          RESPONSE_BIT,
          0x01,
          buildReadIdentityResponse("TTR20", "SN-0000042", "1.2.3"),
          mode,
        ),
      },
      {
        label: "bad checksum",
        bytes: (() => {
          const frame = encodeFrame(0x00, 0x01, buildReadTimeRequest(), mode);
          frame[frame.length - 2] = (frame[frame.length - 2]! ^ 0xff) & 0xff;
          return frame;
        })(),
      },
    ],
    [mode],
  );

  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_22rem]">
      <div className="space-y-4">
        <Panel>
          <PanelHeader
            icon={<Binary className="size-4" />}
            title={dict.protocol.title}
            hint={dict.protocol.subtitle}
            actions={
              <SegmentedControl
                value={mode}
                onChange={setMode}
                options={CHECKSUM_MODES.map((item) => ({ value: item, label: item }))}
              />
            }
          />
          <PanelBody className="space-y-3">
            <Field label={dict.protocol.input} hint={dict.protocol.inputHint}>
              <TextArea
                value={text}
                onChange={(event) => setText(event.target.value)}
                rows={3}
                spellCheck={false}
                className="font-mono text-[0.8125rem] tabular"
              />
            </Field>

            <div className="flex flex-wrap items-center gap-1.5">
              <span className="text-[0.6875rem] text-faint-foreground">{dict.protocol.samples}</span>
              {samples.map((sample) => (
                <Button
                  key={sample.label}
                  size="xs"
                  variant="outline"
                  onClick={() => setText(formatHex(sample.bytes))}
                >
                  {sample.label}
                </Button>
              ))}
              <Button size="xs" variant="ghost" onClick={() => setText("")}>
                {dict.common.clear}
              </Button>
            </div>

            {parsed.invalid.length > 0 || parsed.truncated ? (
              <p className="text-xs text-warning">
                {parsed.invalid.length > 0
                  ? `${parsed.invalid.slice(0, 4).join(", ")}`
                  : dict.protocol.errors.tooShort}
              </p>
            ) : null}

            <div className="grid grid-cols-2 gap-3 border-t border-border pt-3 sm:grid-cols-4">
              <Stat label={dict.protocol.frameLength} value={`${parsed.bytes.length}`} mono />
              <Stat
                label={dict.protocol.payloadLength}
                value={parsed.bytes.length > 1 ? `${parsed.bytes[1]}` : "—"}
                mono
              />
              <Stat label={dict.protocol.mode} value={mode} mono />
              <Stat
                label={dict.protocol.streamFrames}
                value={`${scan.frames.filter((frame) => frame.ok).length}`}
                mono
              />
            </div>
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader
            icon={<Layers className="size-4" />}
            title={dict.protocol.layout}
            hint={dict.protocol.payloadSpan}
          />
          <PanelBody>
            <FrameInspector bytes={parsed.bytes} mode={mode} />
          </PanelBody>
        </Panel>

        {scan.frames.length > 1 ? (
          <Panel>
            <PanelHeader title={dict.protocol.streamTitle} hint={dict.protocol.streamHint} />
            <PanelBody className="space-y-2">
              {scan.frames.map((frame, index) => (
                <div
                  key={index}
                  className={cn(
                    "flex items-center gap-2.5 rounded border px-2.5 py-1.5",
                    frame.ok
                      ? "border-border bg-surface-sunken"
                      : "border-destructive/30 bg-destructive/10",
                  )}
                >
                  <span className="font-mono text-[0.6875rem] text-faint-foreground tabular">
                    #{index + 1}
                  </span>
                  <span className="min-w-0 flex-1 truncate font-mono text-xs text-foreground tabular">
                    {frame.frame ? formatHex(frame.frame.raw) : dict.protocol.invalid}
                  </span>
                  <Badge tone={frame.ok ? "success" : "danger"}>
                    {frame.ok ? dict.protocol.valid : dict.protocol.invalid}
                  </Badge>
                </div>
              ))}
              {scan.rest.length > 0 ? (
                <p className="font-mono text-[0.6875rem] text-faint-foreground">
                  {dict.protocol.streamRest}: {formatHex(scan.rest)}
                </p>
              ) : null}
            </PanelBody>
          </Panel>
        ) : null}
      </div>

      <Panel className="h-fit xl:sticky xl:top-[4.5rem]">
        <PanelHeader
          icon={<Wand2 className="size-4" />}
          title={dict.protocol.builder}
          hint={dict.protocol.builderHint}
        />
        <PanelBody className="space-y-3">
          <Field label={dict.protocol.command}>
            <Select
              value={String(commandId)}
              onChange={(event) => setCommandId(Number(event.target.value))}
            >
              {COMMANDS.map((command) => (
                <option key={command.id} value={command.id}>
                  0x{toHex(command.id)} · {command.name}
                </option>
              ))}
            </Select>
          </Field>

          <div className="grid grid-cols-2 gap-3">
            <Field label={dict.protocol.control}>
              <Select
                value={String(control)}
                onChange={(event) => setControl(Number(event.target.value))}
              >
                {[0x00, 0x01, 0x02, 0x03].map((value) => (
                  <option key={value} value={value}>
                    0x{toHex(value)}
                  </option>
                ))}
              </Select>
            </Field>
            <Field label={dict.protocol.address}>
              <Select
                value={String(address)}
                onChange={(event) => setAddress(Number(event.target.value))}
              >
                {[0x00, 0x01, 0x02, 0x05, 0x0a, 0xff].map((value) => (
                  <option key={value} value={value}>
                    0x{toHex(value)}
                  </option>
                ))}
              </Select>
            </Field>
          </div>

          <SegmentedControl
            value={isResponse ? "response" : "request"}
            onChange={(value) => setIsResponse(value === "response")}
            options={[
              { value: "request", label: dict.common.request },
              { value: "response", label: dict.common.response },
            ]}
            className="w-full"
          />
          <p className="text-xs text-faint-foreground">{dict.protocol.responseBit}</p>

          <div className="rounded border border-border bg-surface-sunken p-2.5">
            <div className="mb-1 text-[0.625rem] tracking-wide text-faint-foreground uppercase">
              {dict.protocol.hex}
            </div>
            <code className="block break-all font-mono text-xs text-primary tabular">
              {formatHex(builderBytes)}
            </code>
          </div>

          <Button className="w-full" size="sm" onClick={() => setText(formatHex(builderBytes))}>
            {dict.protocol.insert}
          </Button>
        </PanelBody>
      </Panel>
    </div>
  );
}
