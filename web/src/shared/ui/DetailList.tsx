import type { ReactNode } from 'react';

export type DetailItem = {
  label: string;
  value: ReactNode;
  mono?: boolean;
};

export function DetailList({ items, className = '' }: { items: DetailItem[]; className?: string }) {
  return (
    <dl className={`detail-list ${className}`}>
      {items.map((item) => (
        <div key={item.label} className="detail-item">
          <dt className="detail-label text-wrap-safe">{item.label}</dt>
          <dd className={`detail-value text-wrap-safe ${item.mono ? 'detail-value--mono' : ''}`}>{item.value}</dd>
        </div>
      ))}
    </dl>
  );
}
