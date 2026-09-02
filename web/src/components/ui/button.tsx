import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import type * as React from "react";

import { cn } from "@/lib/utils";

/*
 * Buttons on an instrument panel.
 *
 * `active:translate-y-px` rather than a scale: a control on a panel presses
 * *in*, it does not shrink. One pixel is enough to feel and small enough that
 * nothing beside it reflows.
 *
 * There is no gradient sheen here. On a consumer surface it reads as polish;
 * next to a table of hex it reads as something moving, and every moving thing
 * on this screen is supposed to mean a frame arrived.
 */
const buttonVariants = cva(
  cn(
    "inline-flex shrink-0 items-center justify-center gap-2 rounded-md text-sm font-medium whitespace-nowrap outline-none",
    "transition-[color,background-color,border-color,box-shadow,translate] duration-150 ease-out active:translate-y-px",
    "focus-visible:ring-[2px] focus-visible:ring-ring/70 focus-visible:ring-offset-2 focus-visible:ring-offset-background",
    "disabled:pointer-events-none disabled:opacity-40",
    "[&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
  ),
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-primary/90",
        outline:
          "border border-border-strong bg-surface text-foreground hover:border-primary/50 hover:bg-surface-raised",
        subtle: "bg-surface-raised text-foreground hover:bg-muted",
        ghost: "text-muted-foreground hover:bg-surface-raised hover:text-foreground",
        danger: "bg-destructive text-destructive-foreground hover:bg-destructive/90",
        success: "bg-success text-success-foreground hover:bg-success/90",
      },
      size: {
        default: "h-9 px-3.5",
        sm: "h-8 px-2.5 text-[0.8125rem]",
        xs: "h-7 gap-1.5 px-2 text-xs [&_svg:not([class*='size-'])]:size-3.5",
        lg: "h-10 px-5",
        icon: "size-9",
        iconSm: "size-8",
      },
    },
    defaultVariants: { variant: "default", size: "default" },
  },
);

export interface ButtonProps
  extends React.ComponentProps<"button">,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
}

export function Button({
  className,
  variant,
  size,
  asChild = false,
  ...props
}: ButtonProps) {
  const Comp = asChild ? Slot : "button";
  return <Comp className={cn(buttonVariants({ variant, size }), className)} {...props} />;
}

export { buttonVariants };
