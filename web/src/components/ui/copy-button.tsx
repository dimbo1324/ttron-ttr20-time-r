"use client";

import { Check, Copy } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";

import { useDictionary } from "@/components/locale-provider";
import { Button } from "@/components/ui/button";

/**
 * Copies a value to the clipboard and says so.
 *
 * Lifting a frame out of the log — into the analyzer, into a ticket, into a
 * message to whoever owns the device — is the most common thing anyone does
 * with one, and selecting spaced hex by hand is exactly the operation a mouse
 * is worst at.
 *
 * The Clipboard API is unavailable over plain HTTP on a non-localhost origin
 * and in older browsers, so a failure is expected rather than exceptional: the
 * button simply does not flip to its confirmed state, and nothing throws into
 * a render.
 */
export function CopyButton({ value, label }: { value: string; label?: string }) {
  const dict = useDictionary();
  const [copied, setCopied] = useState(false);
  const timer = useRef<number | undefined>(undefined);

  // A component unmounted mid-confirmation must not leave a timer holding a
  // setState — the monitor swaps this button out on every selection.
  useEffect(() => () => window.clearTimeout(timer.current), []);

  const copy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      window.clearTimeout(timer.current);
      timer.current = window.setTimeout(() => setCopied(false), 1200);
    } catch {
      setCopied(false);
    }
  }, [value]);

  const title = copied ? dict.common.copied : (label ?? dict.common.copy);

  return (
    <Button
      size="iconSm"
      variant="ghost"
      onClick={copy}
      title={title}
      aria-label={title}
      disabled={value.length === 0}
    >
      {copied ? <Check className="size-4 text-success" /> : <Copy className="size-4" />}
    </Button>
  );
}
