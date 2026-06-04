import { getGatewayStatus, getLastReadTime } from '../../entities/gateway/api';
import { usePollingResource } from '../../shared/lib/usePollingResource';
import { compactNumber, formatDurationMs, formatTime } from '../../shared/lib/format';
import { Badge } from '../../shared/ui/Badge';
import { Card } from '../../shared/ui/Card';
import { PageHeader } from '../../shared/ui/PageHeader';
import { StatCard } from '../../shared/ui/StatCard';
import { ErrorBanner, LoadingState } from '../../shared/ui/State';
import { PollingControlPanel } from '../../widgets/polling-control-panel/PollingControlPanel';

export function GatewayPage() {
  const status = usePollingResource(getGatewayStatus, 2000);
  const lastRead = usePollingResource(getLastReadTime, 2000);
  if (status.loading && !status.data) return <LoadingState label="Loading gateway status" />;
  const data = status.data;
  return (
    <>
      <PageHeader title="Gateway" subtitle="Polling gateway status and control actions." />
      <ErrorBanner message={status.error || lastRead.error} />
      {data ? (
        <div className="mt-4 space-y-5">
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            <StatCard label="Connection" value={data.connected ? 'connected' : data.state} detail={data.targetAddr} />
            <StatCard label="Successful reads" value={compactNumber(data.successfulReads)} detail={`${compactNumber(data.failedReads)} failed`} />
            <StatCard label="Reconnects" value={compactNumber(data.reconnects)} detail={`${compactNumber(data.connectionAttempts)} attempts`} />
            <StatCard label="Device time" value={formatTime(lastRead.data?.deviceTime ?? data.lastDeviceTime)} detail={lastRead.data?.available ? 'available' : 'not available'} />
          </div>
          <Card>
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h2 className="text-lg font-semibold text-zinc-100">Polling session</h2>
              <Badge value={data.connected ? 'connected' : data.state} />
            </div>
            <div className="mt-4 grid gap-3 text-sm text-zinc-400 md:grid-cols-2">
              <div>Checksum mode: <span className="text-zinc-100">{data.checksumMode}</span></div>
              <div>Interval: <span className="text-zinc-100">{formatDurationMs(data.pollingIntervalMs)}</span></div>
              <div>Request timeout: <span className="text-zinc-100">{formatDurationMs(data.requestTimeoutMs)}</span></div>
              <div>Connect timeout: <span className="text-zinc-100">{formatDurationMs(data.connectTimeoutMs)}</span></div>
              <div>Last TX: <span className="text-zinc-100">{formatTime(data.lastTxTime)}</span></div>
              <div>Last RX: <span className="text-zinc-100">{formatTime(data.lastRxTime)}</span></div>
            </div>
          </Card>
          <PollingControlPanel onUpdated={status.refresh} />
        </div>
      ) : null}
    </>
  );
}
