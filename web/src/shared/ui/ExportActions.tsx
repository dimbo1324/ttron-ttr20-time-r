import { Copy, Download, FileJson, Table } from 'lucide-react';
import { useState } from 'react';
import { copyJSON, downloadEndpoint } from '../lib/export';
import { useI18n } from '../i18n/useI18n';
import { Button } from './Button';
import { ActionNotice, ErrorBanner } from './State';

type Props = {
  jsonPath?: string;
  csvPath?: string;
  copyValue?: unknown;
  compact?: boolean;
};

export function ExportActions({ jsonPath, csvPath, copyValue, compact = false }: Props) {
  const { t } = useI18n();
  const [busy, setBusy] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  async function run(label: string, action: () => Promise<void>) {
    setBusy(label);
    setError(null);
    setNotice(null);
    try {
      await action();
      setNotice(label === 'csv' ? t('export.csvStarted') : t('export.jsonStarted'));
    } catch (err) {
      setError(err instanceof Error ? err.message : t('common.exportFailed'));
    } finally {
      setBusy(null);
    }
  }

  async function copy() {
    setBusy('copy');
    setError(null);
    setNotice(null);
    try {
      await copyJSON(copyValue);
      setCopied(true);
      setNotice(t('export.copyNotice'));
      window.setTimeout(() => setCopied(false), 1600);
    } catch {
      setError(t('common.copyFailed'));
    } finally {
      setBusy(null);
    }
  }

  return (
    <div className="flex flex-col gap-2">
      <div className={`button-row ${compact ? 'sm:justify-end' : ''}`}>
        {jsonPath ? (
          <Button
            icon={<FileJson size={16} />}
            tooltip={t('export.jsonTooltip')}
            onClick={() => void run('json', () => downloadEndpoint(jsonPath, 'ft12-export.json'))}
            disabled={busy !== null}
          >
            {t('export.json')}
          </Button>
        ) : null}
        {csvPath ? (
          <Button
            icon={<Table size={16} />}
            tooltip={t('export.csvTooltip')}
            onClick={() => void run('csv', () => downloadEndpoint(csvPath, 'ft12-events.csv'))}
            disabled={busy !== null}
          >
            {t('export.csv')}
          </Button>
        ) : null}
        {copyValue !== undefined ? (
          <Button icon={copied ? <Download size={16} /> : <Copy size={16} />} tooltip={t('export.copyTooltip')} onClick={() => void copy()} disabled={busy !== null}>
            {copied ? t('export.copied') : t('export.copyJson')}
          </Button>
        ) : null}
      </div>
      <ActionNotice message={notice} tone="signal" />
      <ErrorBanner message={error} />
    </div>
  );
}
