export function formatTime(value?: string | null, fallback = 'not available', locale?: string): string {
  if (!value) return fallback;
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString(locale);
}

export function formatDurationMs(ms?: number): string {
  if (ms === undefined || ms === null) return '0 ms';
  if (ms >= 1000) return `${(ms / 1000).toFixed(ms % 1000 === 0 ? 0 : 1)} s`;
  return `${ms} ms`;
}

export function compactNumber(value?: number): string {
  return new Intl.NumberFormat().format(value ?? 0);
}

export function localeForLanguage(language: 'ru' | 'en') {
  return language === 'ru' ? 'ru-RU' : 'en-US';
}
