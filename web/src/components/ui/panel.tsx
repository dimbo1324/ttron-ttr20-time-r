import type * as React from "react";

import { cn } from "@/lib/utils";

/**
 * The one container this UI has.
 *
 * Everything on the bench is a panel: a hairline border, a flat surface and a
 * header strip that names it. There is no second card style and no shadow —
 * on a dark instrument, depth is expensive contrast spent on decoration, and
 * a hairline separates two surfaces just as clearly.
 */
export function Panel({ className, ...props }: React.ComponentProps<"section">) {
  return (
    <section
      className={cn(
        "flex min-w-0 flex-col overflow-hidden rounded-md border border-border bg-card",
        className,
      )}
      {...props}
    />
  );
}

export function PanelHeader({
  className,
  title,
  hint,
  actions,
  icon,
  ...props
}: React.ComponentProps<"header"> & {
  title: React.ReactNode;
  hint?: React.ReactNode;
  actions?: React.ReactNode;
  icon?: React.ReactNode;
}) {
  return (
    <header
      className={cn(
        "flex items-start justify-between gap-3 border-b border-border bg-surface-raised/45 px-3.5 py-2.5",
        className,
      )}
      {...props}
    >
      <div className="flex min-w-0 items-start gap-2.5">
        {icon ? <span className="mt-0.5 text-faint-foreground">{icon}</span> : null}
        <div className="min-w-0">
          <h2 className="truncate text-[0.8125rem] font-semibold text-foreground">{title}</h2>
          {hint ? (
            <p className="mt-0.5 text-xs leading-snug text-faint-foreground">{hint}</p>
          ) : null}
        </div>
      </div>
      {actions ? <div className="flex shrink-0 items-center gap-1.5">{actions}</div> : null}
    </header>
  );
}

export function PanelBody({ className, ...props }: React.ComponentProps<"div">) {
  return <div className={cn("min-w-0 flex-1 p-3.5", className)} {...props} />;
}

/**
 * A labelled value, the atom this whole console is built out of.
 *
 * `mono` switches the value to tabular monospace — used for anything the
 * protocol produced (hex, counters, durations) so digits line up in a column
 * and a changing number does not shuffle its neighbours.
 */
export function Stat({
  label,
  value,
  hint,
  tone = "default",
  mono = false,
  className,
}: {
  label: React.ReactNode;
  value: React.ReactNode;
  hint?: React.ReactNode;
  tone?: "default" | "muted" | "success" | "warning" | "danger" | "primary";
  mono?: boolean;
  className?: string;
}) {
  const toneClass = {
    default: "text-foreground",
    muted: "text-muted-foreground",
    success: "text-success",
    warning: "text-warning",
    danger: "text-destructive",
    primary: "text-primary",
  }[tone];

  return (
    <div className={cn("min-w-0", className)}>
      <div className="truncate text-[0.6875rem] tracking-wide text-faint-foreground uppercase">
        {label}
      </div>
      <div
        className={cn(
          "mt-1 truncate text-[0.9375rem] leading-tight font-medium",
          mono && "font-mono tabular",
          toneClass,
        )}
      >
        {value}
      </div>
      {hint ? <div className="mt-0.5 truncate text-xs text-faint-foreground">{hint}</div> : null}
    </div>
  );
}

/** Key/value row for dense specification lists. */
export function DefRow({
  label,
  value,
  className,
}: {
  label: React.ReactNode;
  value: React.ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "flex items-baseline justify-between gap-3 border-b border-border/60 py-1.5 last:border-0",
        className,
      )}
    >
      <span className="shrink-0 text-xs text-faint-foreground">{label}</span>
      <span className="min-w-0 truncate text-right font-mono text-xs tabular text-foreground">
        {value}
      </span>
    </div>
  );
}
