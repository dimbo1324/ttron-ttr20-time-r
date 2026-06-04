import { useState } from 'react';
import { getPublicConfig } from '../../entities/events/api';
import { usePollingResource } from '../../shared/lib/usePollingResource';
import { Card } from '../../shared/ui/Card';
import { PageHeader } from '../../shared/ui/PageHeader';
import { ErrorBanner, LoadingState } from '../../shared/ui/State';

export function SettingsPage() {
  const config = usePollingResource(getPublicConfig, 10000);
  const [refreshInterval, setRefreshInterval] = useState(2000);
  if (config.loading && !config.data) return <LoadingState label="Loading settings" />;
  return (
    <>
      <PageHeader title="Settings" subtitle="Read-only API wiring and local dashboard preferences." />
      <ErrorBanner message={config.error} />
      <div className="mt-4 grid gap-4 lg:grid-cols-2">
        <Card>
          <h2 className="text-lg font-semibold text-zinc-100">API wiring</h2>
          <div className="mt-4 space-y-3 text-sm text-zinc-400">
            <div>HTTP API endpoint: <span className="text-zinc-100">{import.meta.env.VITE_API_BASE_URL || 'Vite proxy / same origin'}</span></div>
            <div>Emulator gRPC: <span className="text-zinc-100">{config.data?.emulatorGrpc}</span></div>
            <div>Gateway gRPC: <span className="text-zinc-100">{config.data?.gatewayGrpc}</span></div>
            <p className="text-zinc-500">{config.data?.pollingNote}</p>
          </div>
        </Card>
        <Card>
          <h2 className="text-lg font-semibold text-zinc-100">Local refresh</h2>
          <label className="mt-4 block text-sm text-zinc-300">
            Preferred refresh interval ms
            <input className="mt-1 w-full rounded-md border border-line bg-black/20 px-3 py-2 text-zinc-100" type="number" min={1000} step={500} value={refreshInterval} onChange={(event) => setRefreshInterval(Number(event.target.value))} />
          </label>
          <p className="mt-3 text-sm text-zinc-500">Advanced service config remains CLI-based for this milestone.</p>
        </Card>
      </div>
    </>
  );
}
