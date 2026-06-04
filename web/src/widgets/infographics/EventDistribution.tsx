import type { FrameEvent } from '../../entities/events/types';
import { displaySource } from '../../shared/lib/display';
import { useI18n } from '../../shared/i18n/useI18n';
import { Card } from '../../shared/ui/Card';

const directions = ['RX', 'TX', 'ERR', 'SYSTEM'] as const;
const sources = ['emulator', 'gateway'] as const;

export function EventDistribution({ events }: { events: FrameEvent[] }) {
  const { t } = useI18n();
  const byDirection = directions.map((direction) => ({
    label: direction,
    count: events.filter((event) => event.direction === direction).length,
  }));
  const bySource = sources.map((source) => ({
    label: displaySource(source, t),
    count: events.filter((event) => event.source === source || event.service === source).length,
  }));

  return (
    <Card className="h-full">
      <div className="mb-3">
        <h2 className="text-base font-semibold text-ink">{t('infographic.distribution.title')}</h2>
        <p className="text-xs text-subtle">{t('infographic.distribution.subtitle')}</p>
      </div>
      <DistributionGroup title={t('infographic.distribution.byDirection')} items={byDirection} />
      <div className="mt-3">
        <DistributionGroup title={t('infographic.distribution.bySource')} items={bySource} />
      </div>
    </Card>
  );
}

function DistributionGroup({ title, items }: { title: string; items: Array<{ label: string; count: number }> }) {
  const max = Math.max(1, ...items.map((item) => item.count));
  return (
    <div>
      <div className="mb-2 text-xs font-semibold uppercase text-subtle">{title}</div>
      <div className="space-y-2">
        {items.map((item) => (
          <div key={item.label} className="grid grid-cols-[72px_1fr_36px] items-center gap-2 text-xs">
            <div className="truncate font-semibold text-ink">{item.label}</div>
            <div className="h-2 overflow-hidden rounded-full bg-muted">
              <div className="h-full rounded-full bg-signal transition-all" style={{ width: `${Math.max(4, (item.count / max) * 100)}%` }} />
            </div>
            <div className="text-right font-mono text-subtle">{item.count}</div>
          </div>
        ))}
      </div>
    </div>
  );
}
