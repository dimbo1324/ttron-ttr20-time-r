import { getOverview } from '../../entities/events/api';
import { displayChecksum, displayStatus, statTone } from '../../shared/lib/display';
import { usePollingResource } from '../../shared/lib/usePollingResource';
import { compactNumber, formatTime, localeForLanguage } from '../../shared/lib/format';
import { useI18n } from '../../shared/i18n/useI18n';
import { Badge } from '../../shared/ui/Badge';
import { Card } from '../../shared/ui/Card';
import { DetailList } from '../../shared/ui/DetailList';
import { ExportActions } from '../../shared/ui/ExportActions';
import { PageHeader } from '../../shared/ui/PageHeader';
import { StatCard } from '../../shared/ui/StatCard';
import { ErrorBanner, LoadingState } from '../../shared/ui/State';
import { EventDistribution } from '../../widgets/infographics/EventDistribution';
import { ProtocolFlow } from '../../widgets/infographics/ProtocolFlow';
import { RecentEventsTable } from '../../widgets/recent-events-table/RecentEventsTable';

export function DashboardPage() {
  const { t, language } = useI18n();
  const overview = usePollingResource(getOverview, 2000);
  const locale = localeForLanguage(language);
  if (overview.loading && !overview.data) return <LoadingState label={t('common.loadingDashboard')} />;
  const data = overview.data;
  const gatewayState = data?.gateway.connected ? 'connected' : data?.gateway.state;
  const statusItems = data ? [
    { label: t('dashboard.emulatorChecksum'), value: displayChecksum(data.emulator.checksumMode), mono: true },
    { label: t('dashboard.gatewayChecksum'), value: displayChecksum(data.gateway.checksumMode), mono: true },
    { label: t('dashboard.activeConnections'), value: compactNumber(data.emulator.activeConnections) },
    { label: t('dashboard.reconnects'), value: compactNumber(data.gateway.reconnects) },
    { label: t('dashboard.protocolErrors'), value: compactNumber(data.emulator.protocolErrors) },
    { label: t('dashboard.totalEvents'), value: compactNumber(data.events.length) },
  ] : [];

  return (
    <>
      <PageHeader
        title={t('dashboard.title')}
        subtitle={t('dashboard.subtitle')}
        actions={<ExportActions compact jsonPath="/api/v1/export/overview.json" copyValue={data} />}
      />
      <ErrorBanner message={overview.error} />
      {data ? (
        <div className="mt-3 space-y-3">
          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
            <StatCard label={t('dashboard.emulator')} value={displayStatus(data.emulator.state, t)} detail={data.emulator.listenAddr} tone={statTone(data.emulator.state)} />
            <StatCard label={t('dashboard.gateway')} value={displayStatus(gatewayState, t)} detail={data.gateway.targetAddr} tone={statTone(gatewayState)} />
            <StatCard label={t('dashboard.deviceTime')} value={formatTime(data.lastRead.deviceTime, t('common.notAvailable'), locale)} detail={data.lastRead.available ? t('dashboard.lastReadAvailable') : t('dashboard.noReadYet')} tone="signal" />
            <StatCard label={t('dashboard.reads')} value={compactNumber(data.gateway.successfulReads)} detail={t('dashboard.failedReads', { count: compactNumber(data.gateway.failedReads) })} tone={data.gateway.failedReads > 0 ? 'warn' : 'ok'} />
          </div>
          <div className="grid gap-3 xl:grid-cols-[1.45fr_1fr]">
            <ProtocolFlow overview={data} />
            <EventDistribution events={data.events} />
          </div>
          <div className="grid gap-3 xl:grid-cols-[0.9fr_1.4fr]">
            <Card>
              <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
                <h2 className="text-base font-semibold text-ink">{t('dashboard.currentStatus')}</h2>
                <div className="flex gap-2">
                  <Badge value={data.emulator.state} label={displayStatus(data.emulator.state, t)} />
                  <Badge value={gatewayState ?? 'unspecified'} label={displayStatus(gatewayState, t)} />
                </div>
              </div>
              <DetailList items={statusItems} />
            </Card>
            <Card>
              <div className="mb-3 flex items-center justify-between gap-3">
                <h2 className="text-base font-semibold text-ink">{t('dashboard.latestEvents')}</h2>
                {overview.updatedAt ? <span className="text-xs text-subtle">{t('common.updated', { time: overview.updatedAt.toLocaleTimeString(locale) })}</span> : null}
              </div>
              <RecentEventsTable events={data.events} maxHeightClass="max-h-[300px]" />
            </Card>
          </div>
        </div>
      ) : null}
    </>
  );
}
