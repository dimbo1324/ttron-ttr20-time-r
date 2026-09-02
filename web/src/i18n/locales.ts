/**
 * Locale groundwork.
 *
 * Russian is the primary locale — this is a bench for Russian metering
 * equipment and that is who stands in front of it. English is wired in from
 * the start as the second locale: every string on screen goes through a
 * dictionary rather than being written into a component, so adding a third
 * locale later means writing `dictionaries/xx.ts` and listing it here, not
 * touching the UI again.
 */
export const locales = ["ru", "en"] as const;

export type Locale = (typeof locales)[number];

export const defaultLocale: Locale = "ru";

/** Native display name for the switcher — deliberately not translated. */
export const LOCALE_NAMES: Record<Locale, string> = {
  ru: "Русский",
  en: "English",
};

export const LOCALE_SHORT: Record<Locale, string> = {
  ru: "RU",
  en: "EN",
};

/** Where an explicit language choice is remembered, and for how long. */
export const LOCALE_COOKIE = "FT12_LOCALE";
export const LOCALE_COOKIE_MAX_AGE = 60 * 60 * 24 * 365;

export function isLocale(value: string): value is Locale {
  return (locales as readonly string[]).includes(value);
}

/**
 * Reads the locale segment off a pathname (`/en/monitor` → `"en"`), falling
 * back to the default when the path carries no valid prefix. Used by anything
 * that only ever sees a raw path — client components under `usePathname()`.
 */
export function getLocaleFromPathname(pathname: string): Locale {
  const [, segment] = pathname.split("/");
  return segment && isLocale(segment) ? segment : defaultLocale;
}

/** Swaps the locale segment of a path, keeping the rest of the route intact. */
export function withLocale(pathname: string, locale: Locale): string {
  const segments = pathname.split("/").filter((segment) => segment.length > 0);
  if (segments.length > 0 && isLocale(segments[0]!)) segments.shift();
  return `/${[locale, ...segments].join("/")}`;
}
