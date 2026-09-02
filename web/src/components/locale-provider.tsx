"use client";

import { createContext, useContext, type ReactNode } from "react";

import type { Dictionary, Locale } from "@/i18n";

/**
 * Carries the dictionary from the server layout down to client components.
 *
 * The dictionary is resolved once per request on the server and handed over as
 * a plain object, so no locale data is fetched in the browser and there is
 * exactly one place — the `[locale]` layout — that decides which language the
 * tree renders in.
 */
const DictionaryContext = createContext<{ dict: Dictionary; locale: Locale } | null>(null);

export function LocaleProvider({
  dict,
  locale,
  children,
}: {
  dict: Dictionary;
  locale: Locale;
  children: ReactNode;
}) {
  return (
    <DictionaryContext.Provider value={{ dict, locale }}>{children}</DictionaryContext.Provider>
  );
}

export function useDictionary(): Dictionary {
  const context = useContext(DictionaryContext);
  if (!context) throw new Error("useDictionary must be used inside <LocaleProvider>");
  return context.dict;
}

export function useLocale(): Locale {
  const context = useContext(DictionaryContext);
  if (!context) throw new Error("useLocale must be used inside <LocaleProvider>");
  return context.locale;
}
