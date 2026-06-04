export type Theme = 'dark' | 'light';

export const THEME_STORAGE_KEY = 'ft12-ui-theme';

export function isTheme(value: string | null): value is Theme {
  return value === 'dark' || value === 'light';
}
