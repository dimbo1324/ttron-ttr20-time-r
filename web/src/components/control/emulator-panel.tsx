"use client";

import { CircuitBoard, Clock, Zap } from "lucide-react";

import { useDictionary } from "@/components/locale-provider";
import { FrameInspector } from "@/components/protocol/frame-view";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Field, Input, SegmentedControl, SliderRow, ToggleRow } from "@/components/ui/controls";
import { Panel, PanelBody, PanelHeader } from "@/components/ui/panel";
import { formatDriftPerDay, formatDuration, formatPercent } from "@/lib/format";
import { useMounted } from "@/lib/use-mounted";
import {
  buildReadTimeResponse,
  CHECKSUM_MODES,
  encodeFrame,
  RESPONSE_BIT,
  type ChecksumMode,
} from "@/lib/ft12";
import {
  DEFAULT_FAULTS,
  selectActiveFaultCount,
  useBenchStore,
  type DeviceFaults,
} from "@/stores/bench-store";

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
 */

/** Keys into the emulator dictionary section, so a rename cannot slip through. */
type PresetKey =
  | "presetHealthy"
  | "presetNoisyLine"
  | "presetSlowDevice"
  | "presetDriftingClock"
  | "presetDeadDevice";

const PRESETS: { key: PresetKey; faults: Partial<DeviceFaults> }[] = [
  { key: "presetHealthy", faults: { ...DEFAULT_FAULTS } },
  {
    key: "presetNoisyLine",
    faults: { badChecksumProb: 0.35, fragmentProb: 0.25, fragmentDelayMs: 60 },
  },
  { key: "presetSlowDevice", faults: { responseDelayMs: 900, badChecksumProb: 0, noResponse: false } },
  { key: "presetDriftingClock", faults: { clockOffsetMs: 4000, clockDriftPerDayMs: 45_000 } },
  { key: "presetDeadDevice", faults: { noResponse: true } },
];

export function EmulatorPanel() {
  const dict = useDictionary();

  const faults = useBenchStore((state) => state.faults);
  const identity = useBenchStore((state) => state.identity);
  const checksumMode = useBenchStore((state) => state.checksumMode);
  const adapterAddress = useBenchStore((state) => state.adapterAddress);
  const patchFaults = useBenchStore((state) => state.patchFaults);
  const setIdentity = useBenchStore((state) => state.setIdentity);
  const setChecksumMode = useBenchStore((state) => state.setChecksumMode);
  const setAdapterAddress = useBenchStore((state) => state.setAdapterAddress);
  const activeFaults = useBenchStore(selectActiveFaultCount);

  // The preview carries a live timestamp, so it can only be built in the
  // browser: rendering it on the server would bake in a different second.
  const mounted = useMounted();
  const preview = mounted
    ? encodeFrame(
        RESPONSE_BIT,
        adapterAddress,
        buildReadTimeResponse(new Date(Date.now() + faults.clockOffsetMs)),
        checksumMode,
      )
    : [];

  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_24rem]">
      <div className="space-y-4">
        <Panel>
          <PanelHeader
            icon={<Zap className="size-4" />}
            title={dict.emulator.faults}
            hint={dict.emulator.faultsHint}
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
            <SliderRow
              label={dict.emulator.responseDelay}
              hint={dict.emulator.responseDelayHint}
              value={faults.responseDelayMs}
              min={0}
              max={3000}
              step={50}
              format={(value) => (value === 0 ? "0" : formatDuration(value))}
              onChange={(value) => patchFaults({ responseDelayMs: value })}
            />
            <SliderRow
              label={dict.emulator.badChecksum}
              hint={dict.emulator.badChecksumHint}
              value={faults.badChecksumProb}
              min={0}
              max={1}
              step={0.05}
              format={(value) => formatPercent(value, 0)}
              onChange={(value) => patchFaults({ badChecksumProb: value })}
            />
            <SliderRow
              label={dict.emulator.fragment}
              hint={dict.emulator.fragmentHint}
              value={faults.fragmentProb}
              min={0}
              max={1}
              step={0.05}
              format={(value) => formatPercent(value, 0)}
              onChange={(value) => patchFaults({ fragmentProb: value })}
            />
            <SliderRow
              label={dict.emulator.fragmentDelay}
              value={faults.fragmentDelayMs}
              min={0}
              max={800}
              step={20}
              format={(value) => formatDuration(value)}
              onChange={(value) => patchFaults({ fragmentDelayMs: value })}
              disabled={faults.fragmentProb === 0}
            />
            <ToggleRow
              label={dict.emulator.noResponse}
              hint={dict.emulator.noResponseHint}
              checked={faults.noResponse}
              onCheckedChange={(checked) => patchFaults({ noResponse: checked })}
              tone="danger"
            />
            <ToggleRow
              label={dict.emulator.closeAfterRequest}
              hint={dict.emulator.closeAfterRequestHint}
              checked={faults.closeAfterRequest}
              onCheckedChange={(checked) => patchFaults({ closeAfterRequest: checked })}
              tone="danger"
            />
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader
            icon={<Clock className="size-4" />}
            title={dict.emulator.clock}
            hint={dict.emulator.clockHint}
          />
          <PanelBody className="space-y-0">
            <SliderRow
              label={dict.emulator.clockOffset}
              value={faults.clockOffsetMs}
              min={-120_000}
              max={120_000}
              step={1000}
              format={(value) => (value === 0 ? "0" : formatDuration(value, { signed: true }))}
              onChange={(value) => patchFaults({ clockOffsetMs: value })}
            />
            <SliderRow
              label={dict.emulator.clockDrift}
              value={faults.clockDriftPerDayMs}
              min={-300_000}
              max={300_000}
              step={5000}
              format={(value) =>
                value === 0 ? "0" : `${formatDriftPerDay(value)}${dict.common.perDay}`
              }
              onChange={(value) => patchFaults({ clockDriftPerDayMs: value })}
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
            {PRESETS.map((preset) => (
              <Button
                key={preset.key}
                size="xs"
                variant="outline"
                onClick={() => patchFaults({ ...DEFAULT_FAULTS, ...preset.faults })}
              >
                {dict.emulator[preset.key]}
              </Button>
            ))}
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader title={dict.emulator.identity} />
          <PanelBody className="space-y-3">
            <Field label={dict.overview.model}>
              <Input
                value={identity.model}
                onChange={(event) => setIdentity({ model: event.target.value })}
                className="font-mono text-xs"
              />
            </Field>
            <Field label={dict.overview.serial}>
              <Input
                value={identity.serial}
                onChange={(event) => setIdentity({ serial: event.target.value })}
                className="font-mono text-xs"
              />
            </Field>
            <Field label={dict.overview.firmware}>
              <Input
                value={identity.firmware}
                onChange={(event) => setIdentity({ firmware: event.target.value })}
                className="font-mono text-xs"
              />
            </Field>
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader title={dict.overview.checksumMode} />
          <PanelBody className="space-y-3">
            <SegmentedControl<ChecksumMode>
              value={checksumMode}
              onChange={setChecksumMode}
              options={CHECKSUM_MODES.map((mode) => ({ value: mode, label: mode }))}
              className="w-full"
            />
            <Field label={dict.protocol.address}>
              <Input
                type="number"
                min={0}
                max={255}
                value={adapterAddress}
                onChange={(event) => setAdapterAddress(Number(event.target.value))}
                className="font-mono text-xs"
              />
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
