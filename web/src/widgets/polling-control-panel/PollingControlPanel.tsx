import { Play, Square } from 'lucide-react';
import { startGatewayPolling, stopGatewayPolling } from '../../entities/gateway/api';
import { Button } from '../../shared/ui/Button';
import { Card } from '../../shared/ui/Card';

export function PollingControlPanel({ onUpdated }: { onUpdated: () => Promise<void> }) {
  async function start() {
    await startGatewayPolling();
    await onUpdated();
  }
  async function stop() {
    await stopGatewayPolling();
    await onUpdated();
  }
  return (
    <Card>
      <h2 className="text-lg font-semibold text-zinc-100">Polling control</h2>
      <p className="mt-1 text-sm text-zinc-500">Start or stop the gateway polling loop through the HTTP API.</p>
      <div className="mt-4 flex flex-wrap gap-3">
        <Button variant="primary" icon={<Play size={16} />} onClick={() => void start()}>Start polling</Button>
        <Button variant="danger" icon={<Square size={16} />} onClick={() => void stop()}>Stop polling</Button>
      </div>
    </Card>
  );
}
