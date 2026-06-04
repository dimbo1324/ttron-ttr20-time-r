import { useMemo, useState } from 'react';
import { Languages, Moon, Sun } from 'lucide-react';
import { appIcon as AppIcon, routes, type RouteID } from './routes';
import type { Language } from '../shared/i18n/types';
import { useI18n } from '../shared/i18n/useI18n';
import { useTheme } from '../shared/theme/useTheme';

export function App() {
  const { t, language, setLanguage } = useI18n();
  const { theme, toggleTheme } = useTheme();
  const [routeID, setRouteID] = useState<RouteID>('dashboard');
  const active = useMemo(() => routes.find((route) => route.id === routeID) ?? routes[0], [routeID]);
  const Page = active.component;

  return (
    <div className="min-h-screen text-ink">
      <aside className="fixed inset-y-0 left-0 hidden w-56 border-r border-line bg-graphite p-3 md:flex md:flex-col">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg border border-signal/40 bg-signal/10 text-signal">
            <AppIcon size={20} />
          </div>
          <div>
            <div className="font-semibold">{t('app.brand')}</div>
            <div className="text-xs text-subtle">{t('app.subtitle')}</div>
          </div>
        </div>
        <nav className="mt-6 space-y-1.5">
          {routes.map((route) => {
            const Icon = route.icon;
            const selected = active.id === route.id;
            return (
              <button
                key={route.id}
                className={`flex w-full items-center gap-3 rounded-md border px-3 py-2 text-left text-sm transition ${selected ? 'border-signal/50 bg-signal/15 text-signal' : 'border-transparent text-subtle hover:border-line hover:bg-muted hover:text-ink'}`}
                onClick={() => setRouteID(route.id)}
              >
                <Icon size={16} />
                {t(route.labelKey)}
              </button>
            );
          })}
        </nav>
        <div className="mt-auto space-y-2 border-t border-line pt-3">
          <LanguageSwitcher language={language} setLanguage={setLanguage} />
          <button className="flex w-full items-center justify-between rounded-md border border-line bg-muted px-3 py-2 text-sm text-ink transition hover:border-signal" onClick={toggleTheme}>
            <span className="flex items-center gap-2">{theme === 'dark' ? <Moon size={15} /> : <Sun size={15} />}{t('app.theme')}</span>
            <span className="text-xs text-subtle">{theme === 'dark' ? t('app.theme.dark') : t('app.theme.light')}</span>
          </button>
        </div>
      </aside>
      <header className="sticky top-0 z-10 border-b border-line bg-graphite p-3 md:hidden">
        <div className="flex gap-2">
          <select className="app-field min-w-0 flex-1 px-3 py-2" value={routeID} onChange={(event) => setRouteID(event.target.value as RouteID)}>
            {routes.map((route) => <option key={route.id} value={route.id}>{t(route.labelKey)}</option>)}
          </select>
          <button className="app-button app-button--secondary px-2" aria-label={t('app.language')} onClick={() => setLanguage(language === 'ru' ? 'en' : 'ru')}>
            <Languages size={16} />
            <span>{language.toUpperCase()}</span>
          </button>
          <button className="app-button app-button--secondary px-2" aria-label={t('app.theme')} onClick={toggleTheme}>
            {theme === 'dark' ? <Moon size={16} /> : <Sun size={16} />}
          </button>
        </div>
      </header>
      <main className="p-3 md:ml-56 md:p-4 xl:p-5">
        <Page />
      </main>
    </div>
  );
}

function LanguageSwitcher({ language, setLanguage }: { language: Language; setLanguage: (language: Language) => void }) {
  const { t } = useI18n();
  return (
    <div className="rounded-md border border-line bg-muted p-1">
      <div className="mb-1 flex items-center gap-1 px-2 text-xs text-subtle">
        <Languages size={13} />
        {t('app.language')}
      </div>
      <div className="grid grid-cols-2 gap-1">
        {(['ru', 'en'] as const).map((value) => (
          <button
            key={value}
            className={`rounded px-2 py-1.5 text-sm font-semibold transition ${language === value ? 'bg-signal/20 text-signal' : 'text-subtle hover:bg-panel hover:text-ink'}`}
            onClick={() => setLanguage(value)}
          >
            {t(value === 'ru' ? 'app.lang.ru' : 'app.lang.en')}
          </button>
        ))}
      </div>
    </div>
  );
}
