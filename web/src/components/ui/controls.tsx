"use client";

import * as SwitchPrimitive from "@radix-ui/react-switch";
import type * as React from "react";
import { cloneElement, isValidElement, useId } from "react";

import { cn } from "@/lib/utils";

/**
 * Form controls for a settings panel.
 *
 * Every control here is a *labelled row*, not a bare input: on an instrument
 * the label, the value and the reason it exists have to travel together, or
 * the operator has to guess what "0.35" was for. `hint` is part of the control
 * rather than an optional decoration for exactly that reason.
 */

export function Field({
  label,
  hint,
  htmlFor,
  children,
  className,
}: {
  label: React.ReactNode;
  hint?: React.ReactNode;
  htmlFor?: string;
  children: React.ReactNode;
  className?: string;
}) {
  /**
   * The label is bound to its control automatically.
   *
   * Requiring every caller to invent an id and pass it twice is a rule that
   * gets forgotten, and a forgotten one leaves a label pointing at nothing —
   * which reads fine on screen and is invisible to a screen reader. So the
   * field mints an id, hands it to its child, and only steps aside when the
   * caller has supplied one of their own.
   */
  const generated = useId();
  const child = isValidElement(children) ? (children as React.ReactElement<{ id?: string }>) : null;
  const controlId = htmlFor ?? child?.props.id ?? generated;
  const control = child && !htmlFor && !child.props.id ? cloneElement(child, { id: controlId }) : children;

  return (
    <div className={cn("min-w-0 space-y-1.5", className)}>
      <label
        htmlFor={controlId}
        className="block text-[0.6875rem] font-medium tracking-wide text-faint-foreground uppercase"
      >
        {label}
      </label>
      {control}
      {hint ? <p className="text-xs leading-snug text-faint-foreground">{hint}</p> : null}
    </div>
  );
}

export function Input({ className, ...props }: React.ComponentProps<"input">) {
  return (
    <input
      className={cn(
        "h-9 w-full rounded-md border border-input bg-surface-sunken px-2.5 text-sm text-foreground",
        "placeholder:text-faint-foreground",
        "outline-none transition-[border-color,box-shadow] duration-150",
        "focus-visible:border-primary/60 focus-visible:ring-[2px] focus-visible:ring-ring/40",
        "disabled:cursor-not-allowed disabled:opacity-40",
        className,
      )}
      {...props}
    />
  );
}

export function TextArea({ className, ...props }: React.ComponentProps<"textarea">) {
  return (
    <textarea
      className={cn(
        "w-full rounded-md border border-input bg-surface-sunken px-2.5 py-2 text-sm text-foreground",
        "placeholder:text-faint-foreground",
        "outline-none transition-[border-color,box-shadow] duration-150",
        "focus-visible:border-primary/60 focus-visible:ring-[2px] focus-visible:ring-ring/40",
        className,
      )}
      {...props}
    />
  );
}

/**
 * A plain `<select>` rather than a Radix listbox.
 *
 * Native is the right call for short, flat option lists on a dense panel: it
 * is keyboard- and screen-reader-correct with no work, it cannot be clipped by
 * a scroll container, and `color-scheme: dark` on the root already makes the
 * popup match the theme — which was the only reason to reach for a custom one.
 */
export function Select({ className, children, ...props }: React.ComponentProps<"select">) {
  return (
    <div className="relative">
      <select
        className={cn(
          "h-9 w-full appearance-none rounded-md border border-input bg-surface-sunken pr-8 pl-2.5 text-sm text-foreground",
          "outline-none transition-[border-color,box-shadow] duration-150",
          "focus-visible:border-primary/60 focus-visible:ring-[2px] focus-visible:ring-ring/40",
          "disabled:cursor-not-allowed disabled:opacity-40",
          className,
        )}
        {...props}
      >
        {children}
      </select>
      <svg
        aria-hidden
        viewBox="0 0 12 12"
        className="pointer-events-none absolute top-1/2 right-2.5 size-3 -translate-y-1/2 text-faint-foreground"
      >
        <path d="M2 4.5 6 8.5 10 4.5" fill="none" stroke="currentColor" strokeWidth="1.5" />
      </svg>
    </div>
  );
}

export function Switch({
  className,
  ...props
}: React.ComponentProps<typeof SwitchPrimitive.Root>) {
  return (
    <SwitchPrimitive.Root
      className={cn(
        "peer inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border border-border-strong",
        "bg-surface-sunken transition-colors duration-150 outline-none",
        "data-[state=checked]:border-primary/60 data-[state=checked]:bg-primary/80",
        "focus-visible:ring-[2px] focus-visible:ring-ring/50 focus-visible:ring-offset-2 focus-visible:ring-offset-background",
        "disabled:cursor-not-allowed disabled:opacity-40",
        className,
      )}
      {...props}
    >
      <SwitchPrimitive.Thumb
        className={cn(
          "pointer-events-none block size-3.5 translate-x-0.5 rounded-full bg-foreground shadow-sm",
          "transition-transform duration-150 ease-out",
          "data-[state=checked]:translate-x-[1.125rem] data-[state=checked]:bg-primary-foreground",
        )}
      />
    </SwitchPrimitive.Root>
  );
}

/** Switch + label + hint as one row — the shape every fault toggle takes. */
export function ToggleRow({
  label,
  hint,
  checked,
  onCheckedChange,
  disabled,
  tone = "default",
}: {
  label: React.ReactNode;
  hint?: React.ReactNode;
  checked: boolean;
  onCheckedChange: (checked: boolean) => void;
  disabled?: boolean;
  tone?: "default" | "danger";
}) {
  const id = useId();
  return (
    <div
      className={cn(
        "flex items-start justify-between gap-4 border-b border-border/60 py-2.5 last:border-0",
        disabled && "opacity-50",
      )}
    >
      <div className="min-w-0">
        <label
          htmlFor={id}
          className={cn(
            "cursor-pointer text-[0.8125rem] font-medium",
            checked && tone === "danger" ? "text-destructive" : "text-foreground",
          )}
        >
          {label}
        </label>
        {hint ? <p className="mt-0.5 text-xs leading-snug text-faint-foreground">{hint}</p> : null}
      </div>
      <Switch id={id} checked={checked} onCheckedChange={onCheckedChange} disabled={disabled} />
    </div>
  );
}

/**
 * Slider + numeric readout.
 *
 * The readout is not a second input: two editable views of one value is a
 * synchronisation bug waiting to happen, and on a fault panel the operator is
 * dragging to find a threshold, not typing an exact figure.
 */
export function SliderRow({
  label,
  hint,
  value,
  min,
  max,
  step,
  format,
  onChange,
  disabled,
}: {
  label: React.ReactNode;
  hint?: React.ReactNode;
  value: number;
  min: number;
  max: number;
  step: number;
  format: (value: number) => string;
  onChange: (value: number) => void;
  disabled?: boolean;
}) {
  const id = useId();
  return (
    <div className={cn("border-b border-border/60 py-2.5 last:border-0", disabled && "opacity-50")}>
      <div className="flex items-baseline justify-between gap-3">
        <label htmlFor={id} className="text-[0.8125rem] font-medium text-foreground">
          {label}
        </label>
        <span className="font-mono text-xs tabular text-primary">{format(value)}</span>
      </div>
      <input
        id={id}
        type="range"
        min={min}
        max={max}
        step={step}
        value={value}
        disabled={disabled}
        onChange={(event) => onChange(Number(event.target.value))}
        className={cn(
          "mt-2 h-1 w-full cursor-pointer appearance-none rounded-full bg-surface-sunken outline-none",
          "[&::-webkit-slider-thumb]:size-3 [&::-webkit-slider-thumb]:appearance-none",
          "[&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-primary",
          "[&::-moz-range-thumb]:size-3 [&::-moz-range-thumb]:rounded-full",
          "[&::-moz-range-thumb]:border-0 [&::-moz-range-thumb]:bg-primary",
          "focus-visible:ring-[2px] focus-visible:ring-ring/40",
        )}
      />
      {hint ? <p className="mt-1 text-xs leading-snug text-faint-foreground">{hint}</p> : null}
    </div>
  );
}

/** Segmented control — the choice is visible without opening anything. */
export function SegmentedControl<T extends string>({
  value,
  options,
  onChange,
  className,
}: {
  value: T;
  options: { value: T; label: React.ReactNode }[];
  onChange: (value: T) => void;
  className?: string;
}) {
  return (
    <div
      role="radiogroup"
      className={cn(
        "inline-flex items-center gap-0.5 rounded-md border border-border bg-surface-sunken p-0.5",
        className,
      )}
    >
      {options.map((option) => {
        const active = option.value === value;
        return (
          <button
            key={option.value}
            type="button"
            role="radio"
            aria-checked={active}
            onClick={() => onChange(option.value)}
            className={cn(
              "rounded px-2.5 py-1 text-xs font-medium transition-colors duration-150 outline-none",
              "focus-visible:ring-[2px] focus-visible:ring-ring/50",
              active
                ? "bg-surface-raised text-foreground"
                : "text-faint-foreground hover:text-muted-foreground",
            )}
          >
            {option.label}
          </button>
        );
      })}
    </div>
  );
}
