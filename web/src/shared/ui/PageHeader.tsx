import type { ReactNode } from 'react';

export function PageHeader({ title, subtitle, actions }: { title: string; subtitle?: string; actions?: ReactNode }) {
  return (
    <div className="mb-3 flex flex-wrap items-start justify-between gap-3">
      <div className="min-w-0">
        <h1 className="text-wrap-safe text-xl font-semibold leading-tight text-ink md:text-2xl">{title}</h1>
        {subtitle ? <p className="text-wrap-safe mt-1 max-w-3xl text-sm leading-5 text-subtle">{subtitle}</p> : null}
      </div>
      {actions ? <div className="button-row w-full justify-start sm:w-auto sm:justify-end">{actions}</div> : null}
    </div>
  );
}
