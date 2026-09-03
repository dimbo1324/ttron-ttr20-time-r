"use client";

import { CircuitBoard, Clock, Zap } from "lucide-react";

import { SourceNotice } from "@/components/layout/source-notice";
import { useDictionary } from "@/components/locale-provider";
import { FrameInspector } from "@/components/protocol/frame-view";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Field, Input, ReadonlyValue, SegmentedControl, SliderRow, ToggleRow } from "@/components/ui/controls";
import { Panel, PanelBody, PanelHeader } from "@/components/ui/panel";
import {
  buildReadTimeResponse,
  CHECKSUM_MODES,
  encodeFrame,
  RESPONSE_BIT,
  type ChecksumMode,
} from "@/lib/ft12";
import { activeFaultCount, type FaultView } from "@/lib/telemetry/types";
import { useFaultControls, useTelemetry } from "@/lib/telemetry/use-telemetry";
import { useFormat } from "@/lib/use-format";
import { useMounted } from "@/lib/use-mounted";
import { useBenchStore } from "@/stores/bench-store";

/**
 * Emulator control.
 *
 * Every switch here is a fault a real line actually produces, and the panel is
 * ordered the way they escalate: timing first (the device is slow), then
 * integrity (the frame arrives damaged), then availability (it does not arrive
 * at all), then the clock — which is the only fault that leaves the protocol
 * working perfectly while the data is wrong.
 *
 * The live response preview underneath is deliberate: an operator changing the
 * clock offset should see the timestamp inside the frame move, not just a
 * number in a form.
 *
 * ## The same switches, two devices behind them
 *
 * On the bench these drive the device model in this tab. On the live source
 * they drive the real Go emulator over the API, and the panel is the same
 * panel — the point of a bench is that practising on it is practice for the
 * real thing. The one honest difference is the clock section: the Go emulator
 * answers with its host's real time and has no offset to inject, so those two
 * sliders are disabled there and say why.
 */

/** Keys into the emulator dictionary section, so a rename cannot slip through. */
type PresetKey =
  | "presetHealthy"
  | "presetNoisyLine"
  | "presetSlowDevice"
  | "presetDriftingClock"
  | "presetDeadDevice";

/** The state every preset starts from, so one preset cannot leak into the next. */
const CLEARED: Partial<FaultView> = {
  responseDelayMs: 0,
  badChecksumProb: 0,
  fragmentProb: 0,
  fragmentDelayMs: 40,
  noResponse: false,
  closeAfterRequest: false,
  clockOffsetMs: 0,
  clockDriftPerDayMs: 0,
};

const PRESETS: { key: PresetKey; faults: Partial<FaultView>; needsClock?: boolean }[] = [
  { key: "presetHealthy", faults: {} },
  { key: "presetNoisyLine", faults: { badChecksumProb: 0.35, fragmentProb: 0.25, fragmentDelayMs: 60 } },
  { key: "presetSlowDevice", faults: { responseDelayMs: 900 } },
  {
    key: "presetDriftingClock",
    faults: { clockOffsetMs: 4000, clockDriftPerDayMs: 45_000 },
    needsClock: true,
  },
  { key: "presetDeadDevice", faults: { noResponse: true } },
];

export function EmulatorPanel() {
  const dict = useDictionary();
  const format = useFormat();

  const { source, faults, identity: readIdentity, checksumMode: liveMode } = useTelemetry();
  const { setFaults, writable } = useFaultControls();

  // Identity, checksum mode and address describe the device model itself, and
  // only the bench has one to describe: the live emulator is configured by the
  // process that started it.
  const editable = source === "bench";
  const identity = useBenchStore((state) => state.identity);
  const benchMode = useBenchStore((state) => state.checksumMode);
  const adapterAddress = useBenchStore((state) => state.adapterAddress);
  const setIdentity = useBenchStore((state) => state.setIdentity);
  const setChecksumMode = useBenchStore((state) => state.setChecksumMode);
  const setAdapterAddress = useBenchStore((state) => state.setAdapterAddress);

  const checksumMode = editable ? benchMode : liveMode;
  const clockEditable = faults?.clockOffsetMs !== null && faults?.clockOffsetMs !== undefined;
  const activeFaults = activeFaultCount(faults);

  // The preview carries a live timestamp, so it can only be built in the
  // browser: rendering it on the server would bake in a different second.
  const mounted = useMounted();
  const preview = mounted
    ? encodeFrame(
        RESPONSE_BIT,
        adapterAddress,
        buildReadTimeResponse(new Date(Date.now() + (faults?.clockOffsetMs ?? 0))),
        checksumMode,
      )
    : [];

  const disabled = !writable;

  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_24rem]">
      <div className="space-y-4">
        <SourceNotice />

        <Panel>
          <PanelHeader
            icon={<Zap className="size-4" />}
            title={dict.emulator.faults}
            hint={source === "live" ? dict.source.liveFaultsHint : dict.emulator.faultsHint}
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
          <PanelBody className="space-y-0">
            {disabled ? (
              <p className="pb-2 text-xs text-warning">{dict.source.emulatorUnavailable}</p>
            ) : null}
            <SliderRow
              label={dict.emulator.responseDelay}
              hint={dict.emulator.responseDelayHint}
              value={faults?.responseDelayMs ?? 0}
              min={0}
              max={3000}
              step={50}
              format={(value) => (value === 0 ? "0" : format.duration(value))}
              onChange={(value) => setFaults({ responseDelayMs: value })}
              disabled={disabled}
            />
            <SliderRow
              label={dict.emulator.badChecksum}
              hint={dict.emulator.badChecksumHint}
              value={faults?.badChecksumProb ?? 0}
              min={0}
              max={1}
              step={0.05}
              format={(value) => format.percent(value, 0)}
              onChange={(value) => setFaults({ badChecksumProb: value })}
              disabled={disabled}
            />
            <SliderRow
              label={dict.emulator.fragment}
              hint={dict.emulator.fragmentHint}
              value={faults?.fragmentProb ?? 0}
              min={0}
              max={1}
              step={0.05}
              format={(value) => format.percent(value, 0)}
              onChange={(value) => setFaults({ fragmentProb: value })}
              disabled={disabled}
            />
            <SliderRow
              label={dict.emulator.fragmentDelay}
              value={faults?.fragmentDelayMs ?? 0}
              min={0}
              max={800}
              step={20}
              format={(value) => format.duration(value)}
              onChange={(value) => setFaults({ fragmentDelayMs: value })}
              disabled={disabled || !faults?.fragmentProb}
            />
            <ToggleRow
              label={dict.emulator.noResponse}
              hint={dict.emulator.noResponseHint}
              checked={faults?.noResponse ?? false}
              onCheckedChange={(checked) => setFaults({ noResponse: checked })}
              tone="danger"
              disabled={disabled}
            />
            <ToggleRow
              label={dict.emulator.closeAfterRequest}
              hint={dict.emulator.closeAfterRequestHint}
              checked={faults?.closeAfterRequest ?? false}
              onCheckedChange={(checked) => setFaults({ closeAfterRequest: checked })}
              tone="danger"
              disabled={disabled}
            />
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader
            icon={<Clock className="size-4" />}
            title={dict.emulator.clock}
            hint={clockEditable ? dict.emulator.clockHint : dict.source.benchOnlyHint}
            actions={
              clockEditable ? null : <Badge tone="neutral">{dict.source.benchOnly}</Badge>
            }
          />
          <PanelBody className="space-y-0">
            <SliderRow
              label={dict.emulator.clockOffset}
              value={faults?.clockOffsetMs ?? 0}
              min={-120_000}
              max={120_000}
              step={1000}
              format={(value) => (value === 0 ? "0" : format.duration(value, { signed: true }))}
              onChange={(value) => setFaults({ clockOffsetMs: value })}
              disabled={!clockEditable}
            />
            <SliderRow
              label={dict.emulator.clockDrift}
              value={faults?.clockDriftPerDayMs ?? 0}
              min={-300_000}
              max={300_000}
              step={5000}
              format={(value) => (value === 0 ? "0" : format.driftPerDay(value))}
              onChange={(value) => setFaults({ clockDriftPerDayMs: value })}
              disabled={!clockEditable}
            />
          </PanelBody>
        </Panel>
      </div>

      <div className="space-y-4">
        <Panel>
          <PanelHeader
            icon={<CircuitBoard className="size-4" />}
            title={dict.emulator.presets}
            hint={dict.emulator.presetHint}
          />
          <PanelBody className="flex flex-wrap gap-1.5">
            {PRESETS.filter((preset) => clockEditable || !preset.needsClock).map((preset) => (
              <Button
                key={preset.key}
                size="xs"
                variant="outline"
                disabled={disabled}
                onClick={() => setFaults({ ...CLEARED, ...preset.faults })}
              >
                {dict.emulator[preset.key]}
              </Button>
            ))}
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader
            title={dict.emulator.identity}
            hint={editable ? undefined : dict.source.identityLive}
          />
          <PanelBody className="space-y-3">
            <Field label={dict.overview.model}>
              {editable ? (
                <Input
                  value={identity.model}
                  onChange={(event) => setIdentity({ model: event.target.value })}
                  className="font-mono text-xs"
                />
              ) : (
                <ReadonlyValue>{readIdentity?.model ?? dict.common.none}</ReadonlyValue>
              )}
            </Field>
            <Field label={dict.overview.serial}>
              {editable ? (
                <Input
                  value={identity.serial}
                  onChange={(event) => setIdentity({ serial: event.target.value })}
                  className="font-mono text-xs"
                />
              ) : (
                <ReadonlyValue>{readIdentity?.serial ?? dict.common.none}</ReadonlyValue>
              )}
            </Field>
            <Field label={dict.overview.firmware}>
              {editable ? (
                <Input
                  value={identity.firmware}
                  onChange={(event) => setIdentity({ firmware: event.target.value })}
                  className="font-mono text-xs"
                />
              ) : (
                <ReadonlyValue>{readIdentity?.firmware ?? dict.common.none}</ReadonlyValue>
              )}
            </Field>
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader title={dict.overview.checksumMode} />
          <PanelBody className="space-y-3">
            {editable ? (
              <SegmentedControl<ChecksumMode>
                value={checksumMode}
                onChange={setChecksumMode}
                options={CHECKSUM_MODES.map((mode) => ({ value: mode, label: mode }))}
                className="w-full"
              />
            ) : (
              <ReadonlyValue>{checksumMode}</ReadonlyValue>
            )}
            <Field label={dict.protocol.address}>
              {editable ? (
                <Input
                  type="number"
                  min={0}
                  max={255}
                  value={adapterAddress}
                  onChange={(event) => setAdapterAddress(Number(event.target.value))}
                  className="font-mono text-xs"
                />
              ) : (
                <ReadonlyValue>{adapterAddress}</ReadonlyValue>
              )}
            </Field>
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader title={dict.common.response} hint={dict.emulator.clockHint} />
          <PanelBody>
            <FrameInspector bytes={preview} mode={checksumMode} compact />
          </PanelBody>
        </Panel>
      </div>
    </div>
  );
}
