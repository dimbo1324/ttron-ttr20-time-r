import { Card } from './Card';

export function StatCard({ label, value, detail, tone = 'default' }: { label: string; value: string; detail?: string; tone?: 'default' | 'ok' | 'warn' | 'fault' | 'signal' }) {
  const toneClass = {
    default: 'text-ink',
    ok: 'text-ok',
    warn: 'text-warn',
    fault: 'text-fault',
    signal: 'text-signal',
  }[tone];

  return (
    <Card className="min-h-[92px]">
      <div className="text-xs uppercase text-subtle">{label}</div>
      <div className={`mt-1 truncate text-2xl font-semibold ${toneClass}`}>{value}</div>
      {detail ? <div className="mt-1 truncate text-xs text-subtle">{detail}</div> : null}
    </Card>
  );
}
