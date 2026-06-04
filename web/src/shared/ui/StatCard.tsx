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
      <div className="text-wrap-safe text-xs uppercase leading-tight text-subtle" title={label}>{label}</div>
      <div className={`text-wrap-safe mt-1 text-2xl font-semibold leading-tight ${toneClass}`} title={value}>{value}</div>
      {detail ? <div className="text-wrap-safe mt-1 text-xs leading-snug text-subtle" title={detail}>{detail}</div> : null}
    </Card>
  );
}
