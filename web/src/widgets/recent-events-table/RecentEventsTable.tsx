import { Fragment, useState } from 'react';
import type { FrameEvent } from '../../entities/events/types';
import { Badge } from '../../shared/ui/Badge';
import { EmptyState } from '../../shared/ui/State';
import { HexBlock } from '../../shared/ui/HexBlock';
import { formatTime } from '../../shared/lib/format';

export function RecentEventsTable({ events }: { events: FrameEvent[] }) {
  const [expanded, setExpanded] = useState<string | null>(null);
  if (events.length === 0) return <EmptyState label="No recent frames" />;
  return (
    <div className="overflow-hidden rounded-lg border border-line">
      <table className="w-full border-collapse text-left text-sm">
        <thead className="bg-black/20 text-xs uppercase text-zinc-500">
          <tr>
            <th className="px-3 py-2">Time</th>
            <th className="px-3 py-2">Source</th>
            <th className="px-3 py-2">Direction</th>
            <th className="px-3 py-2">Command</th>
            <th className="px-3 py-2">Remote</th>
            <th className="px-3 py-2">Message</th>
          </tr>
        </thead>
        <tbody>
          {events.map((event, index) => {
            const rowKey = `${event.source}-${event.id}-${event.direction}-${event.timestamp}-${event.command ?? ''}-${index}`;

            return (
            <Fragment key={rowKey}>
              <tr
                className="cursor-pointer border-t border-line/70 text-zinc-300 hover:bg-white/5"
                onClick={() => setExpanded(expanded === rowKey ? null : rowKey)}
              >
                <td className="px-3 py-2 text-xs text-zinc-400">{formatTime(event.timestamp)}</td>
                <td className="px-3 py-2">{event.source}</td>
                <td className="px-3 py-2"><Badge value={event.direction} tone={event.direction} /></td>
                <td className="px-3 py-2 font-mono text-xs">{event.command || '-'}</td>
                <td className="px-3 py-2 text-xs">{event.remoteAddr || '-'}</td>
                <td className="px-3 py-2 text-xs text-zinc-400">{event.error || event.message || '-'}</td>
              </tr>
              {expanded === rowKey ? (
                <tr className="border-t border-line/70 bg-black/20">
                  <td colSpan={6} className="px-3 py-3">
                    <HexBlock value={event.rawHex} />
                  </td>
                </tr>
              ) : null}
            </Fragment>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
