export function Toggle({ checked, onChange, label }: { checked: boolean; onChange: (next: boolean) => void; label: string }) {
  return (
    <label className="flex min-h-11 items-center justify-between gap-3 rounded-md border border-line bg-muted px-3 py-2 text-sm text-ink transition hover:border-signal/60">
      <span>{label}</span>
      <input
        type="checkbox"
        className="peer sr-only"
        checked={checked}
        onChange={(event) => onChange(event.target.checked)}
      />
      <span className={`relative h-5 w-9 shrink-0 rounded-full border transition peer-focus-visible:outline peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-signal ${checked ? 'border-signal bg-signal/25' : 'border-line bg-panel'}`}>
        <span className={`absolute left-0.5 top-0.5 h-4 w-4 rounded-full transition-transform ${checked ? 'translate-x-4 bg-signal' : 'bg-subtle'}`} />
      </span>
    </label>
  );
}
