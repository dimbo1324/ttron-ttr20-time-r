"use client";

import { useMemo } from "react";

import { useDictionary, useLocale } from "@/components/locale-provider";
import { createFormatter, type Formatter } from "@/lib/format";

/**
 * The formatters, bound to the active locale and its unit symbols.
 *
 * Components ask for `format.duration(x)` rather than importing a function and
 * remembering to pass the units along; forgetting that argument is what would
 * otherwise leave one panel reading "ms" in a Russian interface.
 */
export function useFormat(): Formatter {
  const dict = useDictionary();
  const locale = useLocale();

  return useMemo(() => createFormatter(locale, dict.units), [locale, dict.units]);
}
