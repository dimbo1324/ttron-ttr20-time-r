import { Fragment, useState } from 'react';
import type { FrameEvent } from '../../entities/events/types';
import { displaySource } from '../../shared/lib/display';
import { formatTime, localeForLanguage } from '../../shared/lib/format';
import { useI18n } from '../../shared/i18n/useI18n';
import { Badge } from '../../shared/ui/Badge';
import { EmptyState } from '../../shared/ui/State';
import { HexBlock } from '../../shared/ui/HexBlock';

export function RecentEventsTable({ events, maxHeightClass = 'max-h-[340px]' }: { events: FrameEvent[]; maxHeightClass?: string }) {
  const { t, language } = useI18n();
  const [expanded, setExpanded] = useState<string | null>(null);
  const locale = localeForLanguage(language);
  if (events.length === 0) return <EmptyState label={t('common.noRecentFrames')} />;
  return (
    <div className={`overflow-auto rounded-md border border-line ${maxHeightClass}`}>
      <table className="responsive-table w-full border-collapse text-left text-sm">
        <colgroup>
          <col className="w-[12rem]" />
          <col className="w-[8rem]" />
          <col className="w-[7rem]" />
          <col className="w-[8rem]" />
          <col className="w-[10rem]" />
          <col />
        </colgroup>
        <thead className="sticky top-0 z-[1] bg-muted text-xs uppercase text-subtle">
          <tr>
            <th className="text-wrap-safe px-3 py-2">{t('table.time')}</th>
            <th className="text-wrap-safe px-3 py-2">{t('table.source')}</th>
            <th className="text-wrap-safe px-3 py-2">{t('table.direction')}</th>
            <th className="text-wrap-safe px-3 py-2">{t('table.command')}</th>
            <th className="text-wrap-safe px-3 py-2">{t('table.remote')}</th>
            <th className="text-wrap-safe px-3 py-2">{t('table.message')}</th>
          </tr>
        </thead>
        <tbody>
          {events.map((event, index) => {
            const rowKey = `${event.source}-${event.id}-${event.direction}-${event.timestamp}-${event.command ?? ''}-${index}`;
            const selected = expanded === rowKey;
            const toggle = () => setExpanded(selected ? null : rowKey);

            return (
            <Fragment key={rowKey}>
              <tr
                aria-expanded={selected}
                aria-label={t('events.expandHint')}
                className="cursor-pointer border-t border-line text-ink transition hover:bg-muted focus-visible:bg-muted"
                onClick={toggle}
                onKeyDown={(keyboardEvent) => {
                  if (keyboardEvent.key === 'Enter' || keyboardEvent.key === ' ') {
                    keyboardEvent.preventDefault();
                    toggle();
                  }
                }}
                role="button"
                tabIndex={0}
              >
                <td className="px-3 py-2 text-xs text-subtle">{formatTime(event.timestamp, t('common.notAvailable'), locale)}</td>
                <td className="text-wrap-safe px-3 py-2">{displaySource(event.source, t)}</td>
                <td className="px-3 py-2"><Badge value={event.direction} tone={event.direction} /></td>
                <td className="text-wrap-safe px-3 py-2 font-mono text-xs">{event.command || '-'}</td>
                <td className="text-wrap-safe px-3 py-2 text-xs">{event.remoteAddr || '-'}</td>
                <td className="px-3 py-2 text-xs text-subtle" title={event.error || event.message || '-'}>
                  <span className="responsive-table__message">{event.error || event.message || '-'}</span>
                </td>
              </tr>
              {selected ? (
                <tr className="border-t border-line bg-muted">
                  <td colSpan={6} className="px-3 py-3">
                    <div className="mb-2 text-xs font-semibold uppercase text-subtle">{t('events.rowDetails')}</div>
                    {event.rawHex ? <HexBlock value={event.rawHex} /> : <EmptyState label={t('events.noRawHex')} />}
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
