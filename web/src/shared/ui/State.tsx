export function ErrorBanner({ message }: { message: string | null }) {
  if (!message) return null;
  return <div className="text-wrap-safe rounded-md border border-fault/40 bg-fault/10 px-3 py-2 text-sm text-fault">{message}</div>;
}

export function ActionNotice({ message, tone = 'ok' }: { message: string | null; tone?: 'ok' | 'signal' | 'warn' }) {
  if (!message) return null;
  return <div className={`action-notice action-notice--${tone}`}>{message}</div>;
}

export function LoadingState({ label = 'Loading' }: { label?: string }) {
  return <div className="app-card text-wrap-safe px-3 py-6 text-center text-sm text-subtle">{label}</div>;
}

export function EmptyState({ label = 'No data' }: { label?: string }) {
  return <div className="text-wrap-safe rounded-md border border-dashed border-line px-3 py-6 text-center text-sm text-subtle">{label}</div>;
}
