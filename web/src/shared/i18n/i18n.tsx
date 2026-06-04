import { useEffect, useMemo, useState, type ReactNode } from 'react';
import { en } from './dictionaries/en';
import { ru } from './dictionaries/ru';
import type { Language, TranslationKey, TranslationValues } from './types';
import { I18nContext, type I18nContextValue } from './I18nContext';

const STORAGE_KEY = 'ft12-ui-language';

const dictionaries = {
  ru,
  en,
} satisfies Record<Language, Record<TranslationKey, string>>;

function initialLanguage(): Language {
  if (typeof window === 'undefined') return 'ru';
  const stored = window.localStorage.getItem(STORAGE_KEY);
  return stored === 'en' || stored === 'ru' ? stored : 'ru';
}

function interpolate(value: string, values?: TranslationValues) {
  if (!values) return value;
  return Object.entries(values).reduce(
    (current, [key, replacement]) => current.replace(new RegExp(`\\{${key}\\}`, 'g'), String(replacement)),
    value,
  );
}

export function I18nProvider({ children }: { children: ReactNode }) {
  const [language, setLanguageState] = useState<Language>(initialLanguage);

  useEffect(() => {
    window.localStorage.setItem(STORAGE_KEY, language);
    document.documentElement.lang = language;
  }, [language]);

  const value = useMemo<I18nContextValue>(() => ({
    language,
    setLanguage: setLanguageState,
    t(key, values) {
      return interpolate(dictionaries[language][key] ?? dictionaries.en[key] ?? key, values);
    },
  }), [language]);

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}
