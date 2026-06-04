import type { ReactNode } from 'react';

export function Card({ children, className = '' }: { children: ReactNode; className?: string }) {
  return <section className={`app-card min-w-0 p-3 transition-shadow duration-150 ${className}`}>{children}</section>;
}
