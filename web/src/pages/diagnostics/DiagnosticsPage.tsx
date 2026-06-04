import { useMemo } from 'react';
import { getHealth, getMetricsText, getReady } from '../../entities/diagnostics/api';
import { getOverview } from '../../entities/events/api';
import { displayStatus, statTone } from '../../shared/lib/display';
import { compactNumber } from '../../shared/lib/format';
import { parseMetricsSummary } from '../../shared/lib/metrics';
import { usePollingResource } from '../../shared/lib/usePollingResource';
import { useI18n } from '../../shared/i18n/useI18n';
import { Badge } from '../../shared/ui/Badge';
import { Card } from '../../shared/ui/Card';
import { ExportActions } from '../../shared/ui/ExportActions';
import { PageHeader } from '../../shared/ui/PageHeader';
import { StatCard } from '../../shared/ui/StatCard';
import { EmptyState, ErrorBanner, LoadingState } from '../../shared/ui/State';
import { EventDistribution } from '../../widgets/infographics/EventDistribution';

export function DiagnosticsPage() {
  const { t } = useI18n();
  const overview = usePollingResource(getOverview, 2000);
  const health = usePollingResource(getHealth, 5000);
  const ready = usePollingResource(getReady, 5000);
  const metrics = usePollingResource(getMetricsText, 5000);
  const metricsSummary = useMemo(() => (metrics.data ? parseMetricsSummary(metrics.data) : null), [metrics.data]);

  if (overview.loading && !overview.data) return <LoadingState label={t('common.loadingDiagnostics')} />;
  const data = overview.data;

  return (
    <>
      <PageHeader
        title={t('diagnostics.title')}
        subtitle={t('diagnostics.subtitle')}
        actions={<ExportActions compact jsonPath="/api/v1/export/overview.json" csvPath="/api/v1/export/events.csv?source=all&limit=100" copyValue={data} />}
      />
      <ErrorBanner message={overview.error || health.error || ready.error || metrics.error} />
      {data ? (
        <div className="mt-3 space-y-3">
          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
            <StatCard label={t('diagnostics.apiHealth')} value={displayStatus(health.data?.status ?? data.health.status, t)} detail={health.data?.service ?? data.health.service} tone={statTone(health.data?.status ?? data.health.status)} />
            <StatCard label={t('diagnostics.readiness')} value={displayStatus(ready.data?.status ?? 'ready', t)} detail={`${t('source.emulator')}: ${ready.data?.emulator ?? 'ok'} / ${t('source.gateway')}: ${ready.data?.gateway ?? 'ok'}`} tone={statTone(ready.data?.status ?? 'ready')} />
            <StatCard label={t('diagnostics.successfulReads')} value={compactNumber(data.gateway.successfulReads)} detail={t('diagnostics.failedReads') + `: ${compactNumber(data.gateway.failedReads)}`} tone={data.gateway.failedReads > 0 ? 'warn' : 'ok'} />
            <StatCard label={t('diagnostics.protocolErrors')} value={compactNumber(data.emulator.protocolErrors)} detail={t('diagnostics.reconnects') + `: ${compactNumber(data.gateway.reconnects)}`} tone={data.emulator.protocolErrors > 0 ? 'fault' : 'ok'} />
          </div>
          <div className="grid gap-3 xl:grid-cols-[0.95fr_1.05fr]">
            <Card>
              <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
                <h2 className="text-base font-semibold text-ink">{t('diagnostics.serviceCounters')}</h2>
                <div className="flex gap-2">
                  <Badge value={data.emulator.state} label={displayStatus(data.emulator.state, t)} />
                  <Badge value={data.gateway.connected ? 'connected' : data.gateway.state} label={displayStatus(data.gateway.connected ? 'connected' : data.gateway.state, t)} />
                </div>
              </div>
              <div className="grid gap-2 text-sm text-subtle sm:grid-cols-2">
                <div>{t('diagnostics.totalRequests')}: <span className="text-ink">{compactNumber(data.emulator.totalRequests)}</span></div>
                <div>{t('diagnostics.totalResponses')}: <span className="text-ink">{compactNumber(data.emulator.totalResponses)}</span></div>
                <div>{t('diagnostics.successfulReads')}: <span className="text-ink">{compactNumber(data.gateway.successfulReads)}</span></div>
                <div>{t('diagnostics.failedReads')}: <span className="text-ink">{compactNumber(data.gateway.failedReads)}</span></div>
                <div>{t('diagnostics.reconnects')}: <span className="text-ink">{compactNumber(data.gateway.reconnects)}</span></div>
                <div>{t('diagnostics.eventsByType')}: <span className="text-ink">{compactNumber(data.events.length)}</span></div>
              </div>
            </Card>
            <EventDistribution events={data.events} />
          </div>
          <div className="grid gap-3 xl:grid-cols-2">
            <Card>
              <h2 className="text-base font-semibold text-ink">{t('diagnostics.metricsSummary')}</h2>
              {metricsSummary ? (
                <div className="mt-3 space-y-2 text-sm text-subtle">
                  <div>{t('diagnostics.totalRequests')}: <span className="text-ink">{compactNumber(metricsSummary.requestsTotal)}</span></div>
                  {metricsSummary.paths.map((item) => (
                    <div key={item.path} className="flex items-center justify-between gap-3 rounded-md border border-line bg-muted px-3 py-2">
                      <span className="truncate font-mono text-xs">{item.path}</span>
                      <span className="font-mono text-xs text-ink">{compactNumber(item.count)}</span>
                    </div>
                  ))}
                </div>
              ) : <EmptyState label={t('diagnostics.metricsUnavailable')} />}
            </Card>
            <Card>
              <h2 className="text-base font-semibold text-ink">{t('diagnostics.docsLinks')}</h2>
              <div className="mt-3 grid gap-2 text-sm text-subtle sm:grid-cols-2">
                <a className="rounded-md border border-line bg-muted px-3 py-2 transition hover:border-signal hover:text-ink" href="https://github.com/dimbo1324/ttron-ttr20-time-r/blob/main/docs/http-api.md">docs/http-api.md</a>
                <a className="rounded-md border border-line bg-muted px-3 py-2 transition hover:border-signal hover:text-ink" href="https://github.com/dimbo1324/ttron-ttr20-time-r/blob/main/docs/web-ui.md">docs/web-ui.md</a>
                <a className="rounded-md border border-line bg-muted px-3 py-2 transition hover:border-signal hover:text-ink" href="https://github.com/dimbo1324/ttron-ttr20-time-r/blob/main/docs/protocol.md">docs/protocol.md</a>
                <a className="rounded-md border border-line bg-muted px-3 py-2 transition hover:border-signal hover:text-ink" href="https://github.com/dimbo1324/ttron-ttr20-time-r/blob/main/docs/examples.md">docs/examples.md</a>
              </div>
              <p className="mt-3 text-xs text-subtle">{t('export.dataNote')}</p>
            </Card>
          </div>
        </div>
      ) : null}
    </>
  );
}
