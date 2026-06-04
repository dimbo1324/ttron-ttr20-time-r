import { Save } from 'lucide-react';
import { useEffect, useState } from 'react';
import type { FaultMode } from '../../entities/emulator/types';
import { updateFaultMode } from '../../entities/emulator/api';
import { useI18n } from '../../shared/i18n/useI18n';
import { Button } from '../../shared/ui/Button';
import { Card } from '../../shared/ui/Card';
import { Toggle } from '../../shared/ui/Toggle';
import { ErrorBanner } from '../../shared/ui/State';

export function FaultModePanel({ faultMode, onUpdated }: { faultMode?: FaultMode; onUpdated: () => Promise<void> }) {
  const { t } = useI18n();
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
      setError(err instanceof Error ? err.message : t('fault.updateFailed'));
    } finally {
      setSaving(false);
    }
  }

  return (
    <Card className="space-y-3">
      <div>
        <h2 className="text-wrap-safe text-base font-semibold text-ink">{t('fault.title')}</h2>
        <p className="text-wrap-safe text-sm text-subtle">{t('fault.subtitle')}</p>
      </div>
      <ErrorBanner message={error} />
      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        <Toggle label={t('fault.corruptChecksum')} checked={draft.corruptChecksum} onChange={(value) => setDraft({ ...draft, corruptChecksum: value, corruptChecksumProbability: value ? Math.max(draft.corruptChecksumProbability, 0.25) : 0 })} />
        <Toggle label={t('fault.fragmentResponse')} checked={draft.fragmentResponse} onChange={(value) => setDraft({ ...draft, fragmentResponse: value, fragmentProbability: value ? Math.max(draft.fragmentProbability, 0.25) : 0 })} />
        <Toggle label={t('fault.noResponse')} checked={draft.noResponse} onChange={(value) => setDraft({ ...draft, noResponse: value })} />
        <Toggle label={t('fault.closeAfterRequest')} checked={draft.closeAfterRequest} onChange={(value) => setDraft({ ...draft, closeAfterRequest: value })} />
      </div>
      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        <label className="text-sm text-ink">
          <span className="text-wrap-safe">{t('fault.responseDelayMs')}</span>
          <input className="app-field mt-1 w-full px-3 py-2" type="number" min={0} value={draft.responseDelayMs} onChange={(event) => setDraft({ ...draft, responseDelayMs: Number(event.target.value) })} />
        </label>
        <label className="text-sm text-ink">
          <span className="text-wrap-safe">{t('fault.fragmentDelayMs')}</span>
          <input className="app-field mt-1 w-full px-3 py-2" type="number" min={0} value={draft.fragmentDelayMs} onChange={(event) => setDraft({ ...draft, fragmentDelayMs: Number(event.target.value) })} />
        </label>
        <label className="text-sm text-ink">
          <span className="text-wrap-safe">{t('fault.corruptProbability')}</span>
          <input className="app-field mt-1 w-full px-3 py-2" type="number" min={0} max={1} step={0.05} value={draft.corruptChecksumProbability} onChange={(event) => setDraft({ ...draft, corruptChecksumProbability: Number(event.target.value) })} />
        </label>
        <label className="text-sm text-ink">
          <span className="text-wrap-safe">{t('fault.fragmentProbability')}</span>
          <input className="app-field mt-1 w-full px-3 py-2" type="number" min={0} max={1} step={0.05} value={draft.fragmentProbability} onChange={(event) => setDraft({ ...draft, fragmentProbability: Number(event.target.value) })} />
        </label>
      </div>
      <div className="button-row">
        <Button variant="primary" icon={<Save size={16} />} onClick={() => void save()} disabled={saving}>
          {saving ? t('common.applying') : t('fault.apply')}
        </Button>
      </div>
    </Card>
  );
}
