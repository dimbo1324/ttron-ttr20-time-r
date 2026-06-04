import { getEmulatorStatus } from '../../entities/emulator/api';
import { usePollingResource } from '../../shared/lib/usePollingResource';
import { displayChecksum, displayStatus, statTone } from '../../shared/lib/display';
import { compactNumber, formatTime, localeForLanguage } from '../../shared/lib/format';
import { useI18n } from '../../shared/i18n/useI18n';
import { Badge } from '../../shared/ui/Badge';
import { Card } from '../../shared/ui/Card';
import { DetailList } from '../../shared/ui/DetailList';
import { ExportActions } from '../../shared/ui/ExportActions';
import { PageHeader } from '../../shared/ui/PageHeader';
import { StatCard } from '../../shared/ui/StatCard';
import { ErrorBanner, LoadingState } from '../../shared/ui/State';
import { FrameAnatomy } from '../../widgets/infographics/FrameAnatomy';
import { FaultModePanel } from '../../widgets/fault-mode-panel/FaultModePanel';

export function EmulatorPage() {
  const { t, language } = useI18n();
  const status = usePollingResource(getEmulatorStatus, 2000);
  const locale = localeForLanguage(language);
  if (status.loading && !status.data) return <LoadingState label={t('common.loadingEmulator')} />;
  const data = status.data;
  const runtimeItems = data ? [
    { label: t('common.checksum'), value: displayChecksum(data.checksumMode), mono: true },
    { label: t('emulator.recentFrames'), value: compactNumber(data.recentFramesCount) },
    { label: t('emulator.lastRequest'), value: formatTime(data.lastRequestTime, t('common.notAvailable'), locale) },
    { label: t('emulator.lastResponse'), value: formatTime(data.lastResponseTime, t('common.notAvailable'), locale) },
  ] : [];
  return (
    <>
      <PageHeader
        title={t('emulator.title')}
        subtitle={t('emulator.subtitle')}
        actions={<ExportActions compact jsonPath="/api/v1/export/emulator-status.json" copyValue={data} />}
      />
      <ErrorBanner message={status.error} />
      {data ? (
        <div className="mt-3 space-y-3">
          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
            <StatCard label={t('emulator.state')} value={displayStatus(data.state, t)} detail={data.listenAddr} tone={statTone(data.state)} />
            <StatCard label={t('emulator.activeConnections')} value={compactNumber(data.activeConnections)} detail={t('emulator.totalConnections', { count: compactNumber(data.totalConnections) })} />
            <StatCard label={t('emulator.requests')} value={compactNumber(data.totalRequests)} detail={t('emulator.responses', { count: compactNumber(data.totalResponses) })} tone="signal" />
            <StatCard label={t('emulator.protocolErrors')} value={compactNumber(data.protocolErrors)} detail={data.lastError || t('emulator.noRecentError')} tone={data.protocolErrors > 0 ? 'fault' : 'ok'} />
          </div>
          <div className="grid gap-3 xl:grid-cols-[0.85fr_1.15fr]">
            <Card>
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h2 className="text-base font-semibold text-ink">{t('emulator.runtime')}</h2>
                <Badge value={data.state} label={displayStatus(data.state, t)} />
              </div>
              <DetailList className="mt-3" items={runtimeItems} />
            </Card>
            <FrameAnatomy checksumMode={data.checksumMode} />
          </div>
          <FaultModePanel faultMode={data.faultMode} onUpdated={status.refresh} />
        </div>
      ) : null}
    </>
  );
}
