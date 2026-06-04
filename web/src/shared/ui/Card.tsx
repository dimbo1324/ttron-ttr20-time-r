import type { ReactNode } from 'react';

export function Card({ children, className = '' }: { children: ReactNode; className?: string }) {
  return <section className={`rounded-lg border border-line bg-panel/90 p-4 shadow-sm ${className}`}>{children}</section>;
}
