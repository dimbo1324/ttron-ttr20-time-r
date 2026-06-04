export function HexBlock({ value }: { value?: string }) {
  if (!value) return null;
  return <pre className="overflow-x-auto rounded-md border border-line bg-muted p-3 font-mono text-xs leading-relaxed text-signal">{value}</pre>;
}
