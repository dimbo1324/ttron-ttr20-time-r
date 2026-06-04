import { Save } from 'lucide-react';
import { useEffect, useState } from 'react';
import type { FaultMode } from '../../entities/emulator/types';
import { updateFaultMode } from '../../entities/emulator/api';
import { Button } from '../../shared/ui/Button';
import { Card } from '../../shared/ui/Card';
import { Toggle } from '../../shared/ui/Toggle';
import { ErrorBanner } from '../../shared/ui/State';

export function FaultModePanel({ faultMode, onUpdated }: { faultMode?: FaultMode; onUpdated: () => Promise<void> }) {
  const [draft, setDraft] = useState<FaultMode | null>(faultMode ?? null);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (faultMode) setDraft(faultMode);
  }, [faultMode]);

  if (!draft) return null;

  async function save() {
    if (!draft) return;
    setSaving(true);
    setError(null);
    try {
      await updateFaultMode(draft);
      await onUpdated();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'fault mode update failed');
    } finally {
      setSaving(false);
    }
  }

  return (
    <Card className="space-y-4">
      <div>
        <h2 className="text-lg font-semibold text-zinc-100">Fault mode</h2>
        <p className="text-sm text-zinc-500">Controls emulator response behavior through the HTTP API.</p>
      </div>
      <ErrorBanner message={error} />
      <div className="grid gap-3 md:grid-cols-2">
        <Toggle label="Corrupt checksum" checked={draft.corruptChecksum} onChange={(value) => setDraft({ ...draft, corruptChecksum: value, corruptChecksumProbability: value ? Math.max(draft.corruptChecksumProbability, 1) : 0 })} />
        <Toggle label="Fragment response" checked={draft.fragmentResponse} onChange={(value) => setDraft({ ...draft, fragmentResponse: value, fragmentProbability: value ? Math.max(draft.fragmentProbability, 1) : 0 })} />
        <Toggle label="No response" checked={draft.noResponse} onChange={(value) => setDraft({ ...draft, noResponse: value })} />
        <Toggle label="Close after request" checked={draft.closeAfterRequest} onChange={(value) => setDraft({ ...draft, closeAfterRequest: value })} />
      </div>
      <div className="grid gap-3 md:grid-cols-2">
        <label className="text-sm text-zinc-300">
          Response delay ms
          <input className="mt-1 w-full rounded-md border border-line bg-black/20 px-3 py-2 text-zinc-100" type="number" min={0} value={draft.responseDelayMs} onChange={(event) => setDraft({ ...draft, responseDelayMs: Number(event.target.value) })} />
        </label>
        <label className="text-sm text-zinc-300">
          Fragment delay ms
          <input className="mt-1 w-full rounded-md border border-line bg-black/20 px-3 py-2 text-zinc-100" type="number" min={0} value={draft.fragmentDelayMs} onChange={(event) => setDraft({ ...draft, fragmentDelayMs: Number(event.target.value) })} />
        </label>
      </div>
      <Button variant="primary" icon={<Save size={16} />} onClick={() => void save()} disabled={saving}>
        {saving ? 'Applying' : 'Apply fault mode'}
      </Button>
    </Card>
  );
}
