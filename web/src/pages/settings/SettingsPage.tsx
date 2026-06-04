import { useEffect, useState } from 'react';
import { getPublicConfig } from '../../entities/events/api';
import { useI18n } from '../../shared/i18n/useI18n';
import { usePollingResource } from '../../shared/lib/usePollingResource';
import { useTheme } from '../../shared/theme/useTheme';
import { Button } from '../../shared/ui/Button';
import { Card } from '../../shared/ui/Card';
import { DetailList } from '../../shared/ui/DetailList';
import { PageHeader } from '../../shared/ui/PageHeader';
import { ErrorBanner, LoadingState } from '../../shared/ui/State';

const REFRESH_STORAGE_KEY = 'ft12-ui-refresh-interval';

export function SettingsPage() {
  const { t, language, setLanguage } = useI18n();
  const { theme, setTheme } = useTheme();
  const config = usePollingResource(getPublicConfig, 10000);
  const [refreshInterval, setRefreshInterval] = useState(() => Number(window.localStorage.getItem(REFRESH_STORAGE_KEY) ?? 2000));

  useEffect(() => {
    window.localStorage.setItem(REFRESH_STORAGE_KEY, String(refreshInterval));
  }, [refreshInterval]);

  if (config.loading && !config.data) return <LoadingState label={t('common.loadingSettings')} />;
  const apiItems = [
    { label: t('settings.httpEndpoint'), value: import.meta.env.VITE_API_BASE_URL || t('settings.viteProxy'), mono: true },
    { label: t('settings.emulatorGrpc'), value: config.data?.emulatorGrpc ?? t('common.notAvailable'), mono: true },
    { label: t('settings.gatewayGrpc'), value: config.data?.gatewayGrpc ?? t('common.notAvailable'), mono: true },
  ];
  return (
    <>
      <PageHeader title={t('settings.title')} subtitle={t('settings.subtitle')} />
      <ErrorBanner message={config.error} />
      <div className="mt-3 grid gap-3 lg:grid-cols-2">
        <Card>
          <h2 className="text-base font-semibold text-ink">{t('settings.apiWiring')}</h2>
          <DetailList className="mt-3" items={apiItems} />
          <p className="text-wrap-safe mt-3 text-sm text-subtle">{config.data?.pollingNote}</p>
        </Card>
        <Card>
          <h2 className="text-base font-semibold text-ink">{t('settings.localPreferences')}</h2>
          <div className="mt-3 grid gap-3 sm:grid-cols-2">
            <div>
              <div className="mb-2 text-sm font-medium text-ink">{t('app.language')}</div>
              <div className="button-row">
                <Button variant={language === 'ru' ? 'primary' : 'secondary'} onClick={() => setLanguage('ru')}>{t('app.lang.ru')}</Button>
                <Button variant={language === 'en' ? 'primary' : 'secondary'} onClick={() => setLanguage('en')}>{t('app.lang.en')}</Button>
              </div>
            </div>
            <div>
              <div className="mb-2 text-sm font-medium text-ink">{t('app.theme')}</div>
              <div className="button-row">
                <Button variant={theme === 'dark' ? 'primary' : 'secondary'} onClick={() => setTheme('dark')}>{t('app.theme.dark')}</Button>
                <Button variant={theme === 'light' ? 'primary' : 'secondary'} onClick={() => setTheme('light')}>{t('app.theme.light')}</Button>
              </div>
            </div>
          </div>
          <label className="mt-4 block text-sm text-ink">
            {t('settings.refreshInterval')}
            <input className="app-field mt-1 w-full px-3 py-2" type="number" min={1000} step={500} value={refreshInterval} onChange={(event) => setRefreshInterval(Number(event.target.value))} />
          </label>
          <p className="mt-3 text-sm text-subtle">{t('settings.refreshNote')}</p>
          <p className="mt-2 text-sm text-subtle">{t('settings.preferencesSaved')}</p>
          <p className="mt-2 text-sm text-subtle">{t('settings.advancedConfig')}</p>
        </Card>
      </div>
    </>
  );
}
