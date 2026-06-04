export function ErrorBanner({ message }: { message: string | null }) {
  if (!message) return null;
  return <div className="rounded-md border border-fault/40 bg-fault/10 px-3 py-2 text-sm text-fault">{message}</div>;
}

export function LoadingState({ label = 'Loading' }: { label?: string }) {
  return <div className="rounded-md border border-line bg-white/5 px-3 py-6 text-center text-sm text-zinc-400">{label}</div>;
}

export function EmptyState({ label = 'No data' }: { label?: string }) {
  return <div className="rounded-md border border-dashed border-line px-3 py-6 text-center text-sm text-zinc-500">{label}</div>;
}
