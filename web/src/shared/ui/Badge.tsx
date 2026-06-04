const tones: Record<string, string> = {
  running: 'app-badge--ok',
  connected: 'app-badge--ok',
  ready: 'app-badge--ok',
  ok: 'app-badge--ok',
  stopped: 'app-badge--neutral',
  unavailable: 'app-badge--neutral',
  unspecified: 'app-badge--neutral',
  degraded: 'app-badge--warn',
  fault: 'app-badge--fault',
  err: 'app-badge--fault',
  error: 'app-badge--fault',
  tx: 'app-badge--signal',
  rx: 'app-badge--ok',
  system: 'app-badge--neutral',
};

export function Badge({ value, tone, label, className = '' }: { value: string; tone?: string; label?: string; className?: string }) {
  const key = (tone ?? value).toLowerCase();
  const text = label ?? value;
  return (
    <span className={`app-badge ${tones[key] ?? 'app-badge--neutral'} ${className}`} title={text}>
      {text}
    </span>
  );
}
