import { createContext } from 'react';
import type { Language, TranslationKey, TranslationValues } from './types';

export type I18nContextValue = {
  language: Language;
  setLanguage: (language: Language) => void;
  t: (key: TranslationKey, values?: TranslationValues) => string;
};

export const I18nContext = createContext<I18nContextValue | null>(null);
