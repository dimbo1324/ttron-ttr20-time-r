import { useEffect, useState } from 'react';
import { getPublicConfig } from '../../entities/events/api';
import { useI18n } from '../../shared/i18n/useI18n';
import { usePollingResource } from '../../shared/lib/usePollingResource';
import { useTheme } from '../../shared/theme/useTheme';
import { Button } from '../../shared/ui/Button';
import { Card } from '../../shared/ui/Card';
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
  return (
    <>
      <PageHeader title={t('settings.title')} subtitle={t('settings.subtitle')} />
      <ErrorBanner message={config.error} />
      <div className="mt-3 grid gap-3 lg:grid-cols-2">
        <Card>
          <h2 className="text-base font-semibold text-ink">{t('settings.apiWiring')}</h2>
          <div className="mt-3 space-y-2 text-sm text-subtle">
            <div>{t('settings.httpEndpoint')}: <span className="text-ink">{import.meta.env.VITE_API_BASE_URL || t('settings.viteProxy')}</span></div>
            <div>{t('settings.emulatorGrpc')}: <span className="text-ink">{config.data?.emulatorGrpc}</span></div>
            <div>{t('settings.gatewayGrpc')}: <span className="text-ink">{config.data?.gatewayGrpc}</span></div>
            <p>{config.data?.pollingNote}</p>
          </div>
        </Card>
        <Card>
          <h2 className="text-base font-semibold text-ink">{t('settings.localPreferences')}</h2>
          <div className="mt-3 grid gap-3 sm:grid-cols-2">
            <div>
              <div className="mb-2 text-sm font-medium text-ink">{t('app.language')}</div>
              <div className="flex gap-2">
                <Button variant={language === 'ru' ? 'primary' : 'secondary'} onClick={() => setLanguage('ru')}>{t('app.lang.ru')}</Button>
                <Button variant={language === 'en' ? 'primary' : 'secondary'} onClick={() => setLanguage('en')}>{t('app.lang.en')}</Button>
              </div>
            </div>
            <div>
              <div className="mb-2 text-sm font-medium text-ink">{t('app.theme')}</div>
              <div className="flex gap-2">
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
