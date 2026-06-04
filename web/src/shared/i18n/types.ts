import type { I18nKey } from './dictionaries/en';

export type Language = 'ru' | 'en';
export type TranslationKey = I18nKey;
export type TranslationValues = Record<string, string | number>;
