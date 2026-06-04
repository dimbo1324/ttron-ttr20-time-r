import { getOverview } from '../../entities/events/api';
import { usePollingResource } from '../../shared/lib/usePollingResource';
import { compactNumber, formatTime } from '../../shared/lib/format';
import { Badge } from '../../shared/ui/Badge';
import { Card } from '../../shared/ui/Card';
import { PageHeader } from '../../shared/ui/PageHeader';
import { StatCard } from '../../shared/ui/StatCard';
import { ErrorBanner, LoadingState } from '../../shared/ui/State';
import { RecentEventsTable } from '../../widgets/recent-events-table/RecentEventsTable';

export function DashboardPage() {
  const overview = usePollingResource(getOverview, 2000);
  if (overview.loading && !overview.data) return <LoadingState label="Loading dashboard" />;
  const data = overview.data;

  return (
    <>
      <PageHeader title="Dashboard" subtitle="Operational view of the FT1.2 emulator and gateway control plane." />
      <ErrorBanner message={overview.error} />
      {data ? (
        <div className="mt-4 space-y-5">
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            <StatCard label="Emulator" value={data.emulator.state} detail={data.emulator.listenAddr} />
            <StatCard label="Gateway" value={data.gateway.connected ? 'connected' : data.gateway.state} detail={data.gateway.targetAddr} />
            <StatCard label="Device time" value={formatTime(data.lastRead.deviceTime)} detail={data.lastRead.available ? 'last read available' : 'no read yet'} />
            <StatCard label="Reads" value={compactNumber(data.gateway.successfulReads)} detail={`${compactNumber(data.gateway.failedReads)} failed`} />
          </div>
          <Card>
            <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
              <h2 className="text-lg font-semibold">Current status</h2>
              <div className="flex gap-2">
                <Badge value={data.emulator.state} />
                <Badge value={data.gateway.connected ? 'connected' : data.gateway.state} />
              </div>
            </div>
            <div className="grid gap-3 text-sm text-zinc-400 md:grid-cols-2">
              <div>Emulator checksum: <span className="text-zinc-100">{data.emulator.checksumMode}</span></div>
              <div>Gateway checksum: <span className="text-zinc-100">{data.gateway.checksumMode}</span></div>
              <div>Active connections: <span className="text-zinc-100">{data.emulator.activeConnections}</span></div>
              <div>Reconnects: <span className="text-zinc-100">{data.gateway.reconnects}</span></div>
            </div>
          </Card>
          <Card>
            <div className="mb-3 flex items-center justify-between">
              <h2 className="text-lg font-semibold">Latest events</h2>
              {overview.updatedAt ? <span className="text-xs text-zinc-500">Updated {overview.updatedAt.toLocaleTimeString()}</span> : null}
            </div>
            <RecentEventsTable events={data.events} />
          </Card>
        </div>
      ) : null}
    </>
  );
}
