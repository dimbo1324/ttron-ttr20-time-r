import { useMemo, useState } from 'react';
import { getEvents } from '../../entities/events/api';
import type { EventDirection, EventSource } from '../../entities/events/types';
import { displaySource } from '../../shared/lib/display';
import { localeForLanguage } from '../../shared/lib/format';
import { usePollingResource } from '../../shared/lib/usePollingResource';
import { useI18n } from '../../shared/i18n/useI18n';
import { Card } from '../../shared/ui/Card';
import { ExportActions } from '../../shared/ui/ExportActions';
import { PageHeader } from '../../shared/ui/PageHeader';
import { ErrorBanner, LoadingState } from '../../shared/ui/State';
import { EventDistribution } from '../../widgets/infographics/EventDistribution';
import { FrameAnatomy } from '../../widgets/infographics/FrameAnatomy';
import { RecentEventsTable } from '../../widgets/recent-events-table/RecentEventsTable';
import { RuntimeTerminal } from '../../widgets/runtime-terminal/RuntimeTerminal';

const directions: Array<'all' | EventDirection> = ['all', 'RX', 'TX', 'ERR', 'SYSTEM'];
const sources: EventSource[] = ['all', 'emulator', 'gateway'];

export function EventsPage() {
  const { t, language } = useI18n();
  const [source, setSource] = useState<EventSource>('all');
  const [direction, setDirection] = useState<'all' | EventDirection>('all');
  const events = usePollingResource(() => getEvents(source, 100), 2000);
  const locale = localeForLanguage(language);
  const filtered = useMemo(() => {
    const rows = events.data?.events ?? [];
    return direction === 'all' ? rows : rows.filter((event) => event.direction === direction);
  }, [events.data, direction]);
  const latestFrame = filtered.find((event) => event.rawHex);
  const query = `source=${source}&limit=100`;

  if (events.loading && !events.data) return <LoadingState label={t('common.loadingEvents')} />;
  return (
    <>
      <PageHeader
        title={t('events.title')}
        subtitle={t('events.subtitle')}
        actions={<ExportActions compact jsonPath={`/api/v1/export/events.json?${query}`} csvPath={`/api/v1/export/events.csv?${query}`} copyValue={{ source, direction, events: filtered }} />}
      />
      <ErrorBanner message={events.error} />
      <div className="mt-3 space-y-3">
        <Card>
          <div className="flex flex-wrap items-end justify-between gap-3">
            <div className="grid flex-1 gap-3 sm:grid-cols-[minmax(11rem,14rem)_minmax(11rem,14rem)_auto]">
              <label className="text-sm text-ink">
                <span className="text-wrap-safe mb-1 block font-medium">{t('common.source')}</span>
                <select className="app-field w-full px-3 py-2" value={source} onChange={(event) => setSource(event.target.value as EventSource)}>
                  {sources.map((value) => <option key={value} value={value}>{displaySource(value, t)}</option>)}
                </select>
              </label>
              <label className="text-sm text-ink">
                <span className="text-wrap-safe mb-1 block font-medium">{t('common.direction')}</span>
                <select className="app-field w-full px-3 py-2" value={direction} onChange={(event) => setDirection(event.target.value as 'all' | EventDirection)}>
                  {directions.map((value) => <option key={value} value={value}>{value === 'all' ? t('common.all') : value}</option>)}
                </select>
              </label>
              <span className="text-wrap-safe self-end rounded-md border border-line bg-muted px-3 py-2 text-sm text-subtle">{t('events.limit100')}</span>
            </div>
            {events.updatedAt ? <span className="text-wrap-safe text-xs text-subtle">{t('common.updated', { time: events.updatedAt.toLocaleTimeString(locale) })}</span> : null}
          </div>
        </Card>
        <div className="grid gap-3 xl:grid-cols-[1fr_1fr]">
          <EventDistribution events={filtered} />
          <FrameAnatomy event={latestFrame} />
        </div>
        <RuntimeTerminal events={filtered} updatedAt={events.updatedAt} maxLines={16} />
        <Card>
          <div className="mb-3 flex items-center justify-between gap-3">
            <h2 className="text-base font-semibold text-ink">{t('events.table')}</h2>
            <span className="text-xs text-subtle">{filtered.length}</span>
          </div>
          <RecentEventsTable events={filtered} maxHeightClass="max-h-[460px]" />
        </Card>
      </div>
    </>
  );
}
