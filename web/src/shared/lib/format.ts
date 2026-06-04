export function formatTime(value?: string | null): string {
  if (!value) return 'not available';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString();
}

export function formatDurationMs(ms?: number): string {
  if (ms === undefined || ms === null) return '0 ms';
  if (ms >= 1000) return `${(ms / 1000).toFixed(ms % 1000 === 0 ? 0 : 1)} s`;
  return `${ms} ms`;
}

export function compactNumber(value?: number): string {
  return new Intl.NumberFormat().format(value ?? 0);
}
