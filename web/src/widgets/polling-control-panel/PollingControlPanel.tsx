import { Play, Square } from 'lucide-react';
import { useState } from 'react';
import { startGatewayPolling, stopGatewayPolling } from '../../entities/gateway/api';
import { useI18n } from '../../shared/i18n/useI18n';
import { Button } from '../../shared/ui/Button';
import { Card } from '../../shared/ui/Card';
import { ErrorBanner } from '../../shared/ui/State';

export function PollingControlPanel({ onUpdated }: { onUpdated: () => Promise<void> }) {
  const { t } = useI18n();
  const [busy, setBusy] = useState<'start' | 'stop' | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function start() {
    setBusy('start');
    setError(null);
    try {
      await startGatewayPolling();
      await onUpdated();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('common.requestFailed'));
    } finally {
      setBusy(null);
    }
  }

  async function stop() {
    setBusy('stop');
    setError(null);
    try {
      await stopGatewayPolling();
      await onUpdated();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('common.requestFailed'));
    } finally {
      setBusy(null);
    }
  }

  return (
    <Card>
      <h2 className="text-base font-semibold text-ink">{t('polling.title')}</h2>
      <p className="mt-1 text-sm text-subtle">{t('polling.subtitle')}</p>
      <ErrorBanner message={error} />
      <div className="mt-3 flex flex-wrap gap-2">
        <Button variant="primary" icon={<Play size={16} />} onClick={() => void start()} disabled={busy !== null}>
          {busy === 'start' ? t('polling.starting') : t('polling.start')}
        </Button>
        <Button variant="danger" icon={<Square size={16} />} onClick={() => void stop()} disabled={busy !== null}>
          {busy === 'stop' ? t('polling.stopping') : t('polling.stop')}
        </Button>
      </div>
    </Card>
  );
}
