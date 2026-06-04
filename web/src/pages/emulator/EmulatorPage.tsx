import { getEmulatorStatus } from '../../entities/emulator/api';
import { usePollingResource } from '../../shared/lib/usePollingResource';
import { compactNumber, formatTime } from '../../shared/lib/format';
import { Badge } from '../../shared/ui/Badge';
import { Card } from '../../shared/ui/Card';
import { PageHeader } from '../../shared/ui/PageHeader';
import { StatCard } from '../../shared/ui/StatCard';
import { ErrorBanner, LoadingState } from '../../shared/ui/State';
import { FaultModePanel } from '../../widgets/fault-mode-panel/FaultModePanel';

export function EmulatorPage() {
  const status = usePollingResource(getEmulatorStatus, 2000);
  if (status.loading && !status.data) return <LoadingState label="Loading emulator status" />;
  const data = status.data;
  return (
    <>
      <PageHeader title="Emulator" subtitle="TCP emulator status and fault-mode controls." />
      <ErrorBanner message={status.error} />
      {data ? (
        <div className="mt-4 space-y-5">
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            <StatCard label="State" value={data.state} detail={data.listenAddr} />
            <StatCard label="Active connections" value={compactNumber(data.activeConnections)} detail={`${compactNumber(data.totalConnections)} total`} />
            <StatCard label="Requests" value={compactNumber(data.totalRequests)} detail={`${compactNumber(data.totalResponses)} responses`} />
            <StatCard label="Protocol errors" value={compactNumber(data.protocolErrors)} detail={data.lastError || 'no recent error'} />
          </div>
          <Card>
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h2 className="text-lg font-semibold text-zinc-100">Runtime</h2>
              <Badge value={data.state} />
            </div>
            <div className="mt-4 grid gap-3 text-sm text-zinc-400 md:grid-cols-2">
              <div>Checksum mode: <span className="text-zinc-100">{data.checksumMode}</span></div>
              <div>Recent frames: <span className="text-zinc-100">{compactNumber(data.recentFramesCount)}</span></div>
              <div>Last request: <span className="text-zinc-100">{formatTime(data.lastRequestTime)}</span></div>
              <div>Last response: <span className="text-zinc-100">{formatTime(data.lastResponseTime)}</span></div>
            </div>
          </Card>
          <FaultModePanel faultMode={data.faultMode} onUpdated={status.refresh} />
        </div>
      ) : null}
    </>
  );
}
