import { cva, type VariantProps } from "class-variance-authority";
import type * as React from "react";

import { cn } from "@/lib/utils";

/**
 * Status pill.
 *
 * Tinted rather than filled: a filled badge is a button-shaped object, and on
 * a panel full of real controls that reads as something to press. A 12% wash
 * of its own hue plus a matching border carries the state at a glance without
 * pretending to be interactive.
 */
const badgeVariants = cva(
  "inline-flex items-center gap-1.5 rounded border px-1.5 py-0.5 text-[0.6875rem] font-medium whitespace-nowrap",
  {
    variants: {
      tone: {
        neutral: "border-border-strong bg-muted/60 text-muted-foreground",
        primary: "border-primary/35 bg-primary/12 text-primary",
        success: "border-success/35 bg-success/12 text-success",
        warning: "border-warning/35 bg-warning/12 text-warning",
        danger: "border-destructive/40 bg-destructive/12 text-destructive",
        tx: "border-signal-tx/35 bg-signal-tx/12 text-signal-tx",
        rx: "border-signal-rx/35 bg-signal-rx/12 text-signal-rx",
        err: "border-signal-err/40 bg-signal-err/12 text-signal-err",
        sys: "border-signal-sys/35 bg-signal-sys/12 text-signal-sys",
      },
      mono: { true: "font-mono tabular", false: "" },
    },
    defaultVariants: { tone: "neutral", mono: false },
  },
);

export function Badge({
  className,
  tone,
  mono,
  ...props
}: React.ComponentProps<"span"> & VariantProps<typeof badgeVariants>) {
  return <span className={cn(badgeVariants({ tone, mono }), className)} {...props} />;
}

export { badgeVariants };
