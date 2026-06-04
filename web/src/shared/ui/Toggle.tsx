export function Toggle({ checked, onChange, label }: { checked: boolean; onChange: (next: boolean) => void; label: string }) {
  return (
    <label className="flex items-center justify-between gap-3 rounded-md border border-line bg-black/15 px-3 py-2 text-sm text-zinc-200">
      <span>{label}</span>
      <input
        type="checkbox"
        className="h-4 w-4 accent-cyan-400"
        checked={checked}
        onChange={(event) => onChange(event.target.checked)}
      />
    </label>
  );
}
