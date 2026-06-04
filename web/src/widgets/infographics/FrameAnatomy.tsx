import type { FrameEvent } from '../../entities/events/types';
import { displayChecksum } from '../../shared/lib/display';
import { useI18n } from '../../shared/i18n/useI18n';
import { Card } from '../../shared/ui/Card';
import { HexBlock } from '../../shared/ui/HexBlock';

export function FrameAnatomy({ event, checksumMode }: { event?: FrameEvent; checksumMode?: string }) {
  const { t } = useI18n();
  const fields = [
    ['0x68', t('infographic.frame.start')],
    ['LEN', t('infographic.frame.length')],
    ['0x68', t('infographic.frame.repeat')],
    ['CONTROL', t('infographic.frame.control')],
    ['ADDRESS', t('infographic.frame.address')],
    ['DATA', t('infographic.frame.data')],
    [displayChecksum(event?.checksumMode ?? checksumMode), t('infographic.frame.checksum')],
    ['0x16', t('infographic.frame.end')],
  ];

  return (
    <Card className="h-full">
      <div className="mb-3">
        <h2 className="text-base font-semibold text-ink">{t('infographic.frame.title')}</h2>
        <p className="text-xs text-subtle">{t('infographic.frame.subtitle')}</p>
      </div>
      <div className="grid grid-cols-2 gap-2 md:grid-cols-4 xl:grid-cols-8">
        {fields.map(([value, label]) => (
          <div key={`${value}-${label}`} className="rounded-md border border-line bg-muted p-2">
            <div className="truncate font-mono text-sm font-semibold text-signal">{value}</div>
            <div className="mt-1 text-xs text-subtle">{label}</div>
          </div>
        ))}
      </div>
      <div className="mt-3">
        <div className="mb-1 text-xs font-semibold uppercase text-subtle">{t('infographic.frame.example')}</div>
        {event?.rawHex ? <HexBlock value={event.rawHex} /> : <HexBlock value="68 03 68 00 01 01 02 16" />}
      </div>
    </Card>
  );
}
