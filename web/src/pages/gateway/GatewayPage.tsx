import { getGatewayStatus, getLastReadTime } from '../../entities/gateway/api';
import { usePollingResource } from '../../shared/lib/usePollingResource';
import { displayChecksum, displayStatus, statTone } from '../../shared/lib/display';
import { compactNumber, formatDurationMs, formatTime, localeForLanguage } from '../../shared/lib/format';
import { useI18n } from '../../shared/i18n/useI18n';
import { Badge } from '../../shared/ui/Badge';
import { Card } from '../../shared/ui/Card';
import { ExportActions } from '../../shared/ui/ExportActions';
import { PageHeader } from '../../shared/ui/PageHeader';
import { StatCard } from '../../shared/ui/StatCard';
import { ErrorBanner, LoadingState } from '../../shared/ui/State';
import { PollingTimeline } from '../../widgets/infographics/PollingTimeline';
import { PollingControlPanel } from '../../widgets/polling-control-panel/PollingControlPanel';

export function GatewayPage() {
  const { t, language } = useI18n();
  const status = usePollingResource(getGatewayStatus, 2000);
  const lastRead = usePollingResource(getLastReadTime, 2000);
  const locale = localeForLanguage(language);
  if (status.loading && !status.data) return <LoadingState label={t('common.loadingGateway')} />;
  const data = status.data;
  const gatewayState = data?.connected ? 'connected' : data?.state;
  return (
    <>
      <PageHeader
        title={t('gateway.title')}
        subtitle={t('gateway.subtitle')}
        actions={<ExportActions compact jsonPath="/api/v1/export/gateway-status.json" copyValue={data} />}
      />
      <ErrorBanner message={status.error || lastRead.error} />
      {data ? (
        <div className="mt-3 space-y-3">
          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
            <StatCard label={t('gateway.connection')} value={displayStatus(gatewayState, t)} detail={data.targetAddr} tone={statTone(gatewayState)} />
            <StatCard label={t('gateway.successfulReads')} value={compactNumber(data.successfulReads)} detail={t('dashboard.failedReads', { count: compactNumber(data.failedReads) })} tone={data.failedReads > 0 ? 'warn' : 'ok'} />
            <StatCard label={t('gateway.reconnects')} value={compactNumber(data.reconnects)} detail={t('gateway.attempts', { count: compactNumber(data.connectionAttempts) })} />
            <StatCard label={t('gateway.deviceTime')} value={formatTime(lastRead.data?.deviceTime ?? data.lastDeviceTime, t('common.notAvailable'), locale)} detail={lastRead.data?.available ? t('common.available') : t('common.notAvailable')} tone="signal" />
          </div>
          <div className="grid gap-3 xl:grid-cols-[0.85fr_1.15fr]">
            <Card>
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h2 className="text-base font-semibold text-ink">{t('gateway.pollingSession')}</h2>
                <Badge value={gatewayState ?? 'unspecified'} label={displayStatus(gatewayState, t)} />
              </div>
              <div className="mt-3 grid gap-2 text-sm text-subtle md:grid-cols-2 xl:grid-cols-1">
                <div>{t('common.checksum')}: <span className="text-ink">{displayChecksum(data.checksumMode)}</span></div>
                <div>{t('gateway.interval')}: <span className="text-ink">{formatDurationMs(data.pollingIntervalMs)}</span></div>
                <div>{t('gateway.requestTimeout')}: <span className="text-ink">{formatDurationMs(data.requestTimeoutMs)}</span></div>
                <div>{t('gateway.connectTimeout')}: <span className="text-ink">{formatDurationMs(data.connectTimeoutMs)}</span></div>
                <div>{t('gateway.lastTx')}: <span className="text-ink">{formatTime(data.lastTxTime, t('common.notAvailable'), locale)}</span></div>
                <div>{t('gateway.lastRx')}: <span className="text-ink">{formatTime(data.lastRxTime, t('common.notAvailable'), locale)}</span></div>
                <div>{t('gateway.lastError')}: <span className="text-ink">{data.lastError || t('common.none')}</span></div>
              </div>
            </Card>
            <PollingTimeline status={data} />
          </div>
          <PollingControlPanel onUpdated={status.refresh} />
        </div>
      ) : null}
    </>
  );
}
