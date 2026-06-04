import { useMemo, useState } from 'react';
import { getEvents } from '../../entities/events/api';
import type { EventDirection, EventSource } from '../../entities/events/types';
import { usePollingResource } from '../../shared/lib/usePollingResource';
import { Card } from '../../shared/ui/Card';
import { PageHeader } from '../../shared/ui/PageHeader';
import { ErrorBanner, LoadingState } from '../../shared/ui/State';
import { RecentEventsTable } from '../../widgets/recent-events-table/RecentEventsTable';

const directions: Array<'all' | EventDirection> = ['all', 'RX', 'TX', 'ERR', 'SYSTEM'];

export function EventsPage() {
  const [source, setSource] = useState<EventSource>('all');
  const [direction, setDirection] = useState<'all' | EventDirection>('all');
  const events = usePollingResource(() => getEvents(source, 100), 2000);
  const filtered = useMemo(() => {
    const rows = events.data?.events ?? [];
    return direction === 'all' ? rows : rows.filter((event) => event.direction === direction);
  }, [events.data, direction]);

  if (events.loading && !events.data) return <LoadingState label="Loading events" />;
  return (
    <>
      <PageHeader title="Events / Frames" subtitle="Recent emulator and gateway frames with expandable raw hex." />
      <ErrorBanner message={events.error} />
      <Card className="mt-4">
        <div className="mb-4 flex flex-wrap gap-3">
          <label className="text-sm text-zinc-300">
            Source
            <select className="ml-2 rounded-md border border-line bg-black/20 px-3 py-2" value={source} onChange={(event) => setSource(event.target.value as EventSource)}>
              <option value="all">all</option>
              <option value="emulator">emulator</option>
              <option value="gateway">gateway</option>
            </select>
          </label>
          <label className="text-sm text-zinc-300">
            Direction
            <select className="ml-2 rounded-md border border-line bg-black/20 px-3 py-2" value={direction} onChange={(event) => setDirection(event.target.value as 'all' | EventDirection)}>
              {directions.map((value) => <option key={value} value={value}>{value}</option>)}
            </select>
          </label>
          {events.updatedAt ? <span className="self-center text-xs text-zinc-500">Updated {events.updatedAt.toLocaleTimeString()}</span> : null}
        </div>
        <RecentEventsTable events={filtered} />
      </Card>
    </>
  );
}
