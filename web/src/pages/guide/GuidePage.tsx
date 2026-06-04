import { BookOpen, ShieldAlert } from 'lucide-react';
import { useI18n } from '../../shared/i18n/useI18n';
import type { TranslationKey } from '../../shared/i18n/types';
import { Card } from '../../shared/ui/Card';
import { PageHeader } from '../../shared/ui/PageHeader';
import { FrameAnatomy } from '../../widgets/infographics/FrameAnatomy';
import { ProtocolFlow } from '../../widgets/infographics/ProtocolFlow';

const sections: Array<{ title: TranslationKey; body: TranslationKey }> = [
  { title: 'guide.project.title', body: 'guide.project.body' },
  { title: 'guide.protocol.title', body: 'guide.protocol.body' },
  { title: 'guide.emulator.title', body: 'guide.emulator.body' },
  { title: 'guide.gateway.title', body: 'guide.gateway.body' },
  { title: 'guide.api.title', body: 'guide.api.body' },
  { title: 'guide.web.title', body: 'guide.web.body' },
  { title: 'guide.events.title', body: 'guide.events.body' },
  { title: 'guide.checksum.title', body: 'guide.checksum.body' },
  { title: 'guide.faults.title', body: 'guide.faults.body' },
  { title: 'guide.polling.title', body: 'guide.polling.body' },
  { title: 'guide.exports.title', body: 'guide.exports.body' },
];

export function GuidePage() {
  const { t } = useI18n();

  return (
    <>
      <PageHeader title={t('guide.title')} subtitle={t('guide.subtitle')} />
      <div className="mt-3 space-y-3">
        <div className="grid gap-3 xl:grid-cols-2">
          <ProtocolFlow />
          <FrameAnatomy />
        </div>
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          {sections.map((section) => (
            <Card key={section.title}>
              <div className="mb-2 flex items-center gap-2">
                <BookOpen className="shrink-0 text-signal" size={16} />
                <h2 className="text-wrap-safe text-base font-semibold text-ink">{t(section.title)}</h2>
              </div>
              <p className="text-wrap-safe text-sm leading-6 text-subtle">{t(section.body)}</p>
            </Card>
          ))}
          <Card className="border-fault/40">
            <div className="mb-2 flex items-center gap-2">
              <ShieldAlert className="shrink-0 text-fault" size={16} />
              <h2 className="text-wrap-safe text-base font-semibold text-ink">{t('guide.safety.title')}</h2>
            </div>
            <p className="text-wrap-safe text-sm leading-6 text-subtle">{t('guide.safety.body')}</p>
          </Card>
        </div>
      </div>
    </>
  );
}
