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

export function Badge({ value, tone, label }: { value: string; tone?: string; label?: string }) {
  const key = (tone ?? value).toLowerCase();
  return (
    <span className={`app-badge ${tones[key] ?? 'app-badge--neutral'}`}>
      {label ?? value}
    </span>
  );
}
