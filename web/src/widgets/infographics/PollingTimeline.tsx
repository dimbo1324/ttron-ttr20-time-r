import { AlertTriangle, CheckCircle2, Circle, RadioTower } from 'lucide-react';
import type { GatewayStatus } from '../../entities/gateway/types';
import { formatTime, localeForLanguage } from '../../shared/lib/format';
import { useI18n } from '../../shared/i18n/useI18n';
import { Card } from '../../shared/ui/Card';

export function PollingTimeline({ status }: { status?: GatewayStatus | null }) {
  const { t, language } = useI18n();
  const locale = localeForLanguage(language);
  const steps = [
    { label: t('infographic.timeline.connect'), detail: status?.targetAddr, done: (status?.connectionAttempts ?? 0) > 0 },
    { label: t('infographic.timeline.tx'), detail: formatTime(status?.lastTxTime, t('common.notAvailable'), locale), done: Boolean(status?.lastTxTime) },
    { label: t('infographic.timeline.rx'), detail: formatTime(status?.lastRxTime, t('common.notAvailable'), locale), done: Boolean(status?.lastRxTime) },
    { label: t('infographic.timeline.parse'), detail: `${status?.recentFramesCount ?? 0}`, done: (status?.recentFramesCount ?? 0) > 0 },
    { label: t('infographic.timeline.status'), detail: status?.connected ? t('status.connected') : t('status.stopped'), done: Boolean(status?.connected) },
    { label: t('infographic.timeline.retry'), detail: status?.lastError || t('common.none'), done: Boolean(status?.lastError || (status?.failedReads ?? 0) > 0), error: Boolean(status?.lastError) },
  ];

  return (
    <Card className="h-full">
      <div className="mb-3 flex items-start justify-between gap-3">
        <div>
          <h2 className="text-base font-semibold text-ink">{t('infographic.timeline.title')}</h2>
          <p className="text-xs text-subtle">{t('infographic.timeline.subtitle')}</p>
        </div>
        <RadioTower className="text-signal" size={18} />
      </div>
      <div className="grid gap-2 md:grid-cols-3 xl:grid-cols-6">
        {steps.map((step) => {
          const Icon = step.error ? AlertTriangle : step.done ? CheckCircle2 : Circle;
          return (
            <div key={step.label} className="rounded-md border border-line bg-muted p-2">
              <div className="flex items-center gap-2">
                <Icon className={step.error ? 'text-fault' : step.done ? 'text-ok' : 'text-subtle'} size={16} />
                <span className="text-sm font-semibold text-ink">{step.label}</span>
              </div>
              <div className="mt-2 truncate text-xs text-subtle">{step.detail || t('common.notAvailable')}</div>
            </div>
          );
        })}
      </div>
    </Card>
  );
}
