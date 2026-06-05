import { Activity, CircleDot, Terminal } from 'lucide-react';
import type { FrameEvent } from '../../entities/events/types';
import { useI18n } from '../../shared/i18n/useI18n';
import { formatTime, localeForLanguage } from '../../shared/lib/format';
import { Card } from '../../shared/ui/Card';

type Props = {
  events: FrameEvent[];
  updatedAt?: Date | null;
  maxLines?: number;
};

const directionTone: Record<string, string> = {
  RX: 'text-ok',
  TX: 'text-signal',
  ERR: 'text-fault',
  SYSTEM: 'text-warn',
};

export function RuntimeTerminal({ events, updatedAt, maxLines = 12 }: Props) {
  const { t, language } = useI18n();
  const locale = localeForLanguage(language);
  const lines = events.slice(0, maxLines).map((event) => formatLine(event, locale, t('common.notAvailable')));
  const ticker = lines.length > 0 ? lines.slice(0, 6).map((line) => `${line.time} ${line.text}`).join('  //  ') : t('terminal.waiting');

  return (
    <Card className="runtime-terminal p-0">
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-line px-3 py-2">
        <div className="flex min-w-0 items-center gap-2">
          <Terminal className="shrink-0 text-signal" size={17} />
          <div className="min-w-0">
            <h2 className="text-wrap-safe text-base font-semibold text-ink">{t('terminal.title')}</h2>
            <p className="text-wrap-safe text-xs text-subtle">{t('terminal.subtitle')}</p>
          </div>
        </div>
        <div className="flex items-center gap-2 rounded-md border border-line bg-muted px-2 py-1 text-xs text-subtle">
          <Activity className="pulse-dot shrink-0 text-ok" size={12} />
          <span className="text-wrap-safe">{updatedAt ? t('common.updated', { time: updatedAt.toLocaleTimeString(locale) }) : t('terminal.waiting')}</span>
        </div>
      </div>
      <div className="runtime-terminal__ticker" aria-label={t('terminal.tickerLabel')}>
        <div className="runtime-terminal__track">
          <span>{ticker}</span>
          <span aria-hidden="true">{ticker}</span>
        </div>
      </div>
      <div className="runtime-terminal__body">
        {lines.length > 0 ? (
          lines.map((line, index) => (
            <div className="runtime-terminal__line" key={`${line.id}-${index}`}>
              <span className="runtime-terminal__time">{line.time}</span>
              <span className={`runtime-terminal__direction ${directionTone[line.direction] ?? 'text-subtle'}`}>
                <CircleDot size={10} />
                {line.direction}
              </span>
              <span className="runtime-terminal__text">{line.text}</span>
            </div>
          ))
        ) : (
          <div className="runtime-terminal__empty">{t('terminal.noEvents')}</div>
        )}
      </div>
    </Card>
  );
}

function formatLine(event: FrameEvent, locale: string, empty: string) {
  const time = formatTime(event.timestamp, empty, locale);
  const command = event.command || event.message || event.error || 'frame';
  const remote = event.remoteAddr ? ` @ ${event.remoteAddr}` : '';
  const text = `${event.source || event.service || 'service'} ${event.direction} ${command}${remote}`;
  return { id: event.id, direction: event.direction, text, time };
}
