import type { Overview } from '../../entities/events/types';
import { displayStatus } from '../../shared/lib/display';
import { Badge } from '../../shared/ui/Badge';
import { Card } from '../../shared/ui/Card';
import { InfoTile } from '../../shared/ui/InfoTile';
import { useI18n } from '../../shared/i18n/useI18n';

export function ProtocolFlow({ overview }: { overview?: Overview | null }) {
  const { t } = useI18n();
  const nodes = [
    { label: t('infographic.protocolFlow.web'), detail: t('infographic.protocolFlow.controlPlane'), tone: 'running', status: t('status.running') },
    { label: t('infographic.protocolFlow.http'), detail: overview?.health.service ?? 'ft12-api', tone: overview?.health.status ?? 'ok', status: displayStatus(overview?.health.status ?? 'ok', t) },
    { label: t('infographic.protocolFlow.grpc'), detail: t('infographic.protocolFlow.controlPlane'), tone: overview ? 'connected' : 'unavailable', status: overview ? t('status.connected') : t('status.unavailable') },
    { label: t('infographic.protocolFlow.gateway'), detail: t('infographic.protocolFlow.dataPath'), tone: overview?.gateway.connected ? 'connected' : overview?.gateway.state, status: displayStatus(overview?.gateway.connected ? 'connected' : overview?.gateway.state, t) },
    { label: t('infographic.protocolFlow.ft12'), detail: 'TCP 9000', tone: overview?.gateway.connected ? 'connected' : 'stopped', status: overview?.gateway.connected ? t('status.connected') : t('status.stopped') },
    { label: t('infographic.protocolFlow.emulator'), detail: overview?.emulator.listenAddr ?? 'TCP', tone: overview?.emulator.state, status: displayStatus(overview?.emulator.state, t) },
  ];

  return (
    <Card className="h-full">
      <div className="mb-3 flex items-start justify-between gap-3">
        <div>
          <h2 className="text-base font-semibold text-ink">{t('infographic.protocolFlow.title')}</h2>
          <p className="text-xs text-subtle">{t('infographic.protocolFlow.subtitle')}</p>
        </div>
      </div>
      <div className="info-grid">
        {nodes.map((node, index) => (
          <InfoTile
            key={`${index}-${node.label}`}
            title={node.label}
            detail={node.detail}
            icon={<span className="flex h-5 w-5 items-center justify-center rounded-full border border-line bg-panel text-[11px] font-semibold text-subtle">{index + 1}</span>}
            badge={<Badge value={node.tone ?? 'unspecified'} label={node.status} />}
          />
        ))}
      </div>
    </Card>
  );
}
