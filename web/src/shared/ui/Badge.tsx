const tones: Record<string, string> = {
  running: 'border-ok/40 bg-ok/10 text-ok',
  connected: 'border-ok/40 bg-ok/10 text-ok',
  stopped: 'border-zinc-500/40 bg-zinc-500/10 text-zinc-300',
  degraded: 'border-warn/40 bg-warn/10 text-warn',
  fault: 'border-fault/40 bg-fault/10 text-fault',
  err: 'border-fault/40 bg-fault/10 text-fault',
  tx: 'border-signal/40 bg-signal/10 text-signal',
  rx: 'border-ok/40 bg-ok/10 text-ok',
};

export function Badge({ value, tone }: { value: string; tone?: string }) {
  const key = (tone ?? value).toLowerCase();
  return (
    <span className={`inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-medium ${tones[key] ?? 'border-line bg-white/5 text-zinc-300'}`}>
      {value}
    </span>
  );
}
