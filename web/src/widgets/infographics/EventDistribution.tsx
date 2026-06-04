import type { FrameEvent } from '../../entities/events/types';
import { displaySource } from '../../shared/lib/display';
import { countEventsByDirection, countEventsBySource, type EventStat } from '../../shared/lib/eventStats';
import { useI18n } from '../../shared/i18n/useI18n';
import { Card } from '../../shared/ui/Card';

export function EventDistribution({ events }: { events: FrameEvent[] }) {
  const { t } = useI18n();
  const byDirection = countEventsByDirection(events);
  const bySource = countEventsBySource(events, (source) => displaySource(source, t));

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

function DistributionGroup({ title, items }: { title: string; items: EventStat[] }) {
  const max = Math.max(1, ...items.map((item) => item.count));
  return (
    <div>
      <div className="mb-2 text-xs font-semibold uppercase text-subtle">{title}</div>
      <div className="space-y-2">
        {items.map((item) => (
          <div key={item.label} className="grid grid-cols-[minmax(5rem,0.8fr)_minmax(5rem,1fr)_2.5rem] items-center gap-2 text-xs">
            <div className="text-wrap-safe font-semibold text-ink">{item.label}</div>
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
