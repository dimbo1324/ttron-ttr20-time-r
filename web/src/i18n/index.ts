import { en } from "./dictionaries/en";
import { ru } from "./dictionaries/ru";
import {
  defaultLocale,
  getLocaleFromPathname,
  isLocale,
  LOCALE_COOKIE,
  LOCALE_COOKIE_MAX_AGE,
  LOCALE_NAMES,
  LOCALE_SHORT,
  locales,
  withLocale,
  type Locale,
} from "./locales";

export {
  defaultLocale,
  getLocaleFromPathname,
  isLocale,
  LOCALE_COOKIE,
  LOCALE_COOKIE_MAX_AGE,
  LOCALE_NAMES,
  LOCALE_SHORT,
  locales,
  withLocale,
};
export type { Locale };

/**
 * Deep-widens literal types back to their base primitive.
 *
 * `ru` is authored `as const`, which narrows every string to its exact Russian
 * wording. Used directly as `Dictionary` that would reject `en` for the sole
 * reason that it says something else — this widens the *values* while keeping
 * the *shape*, which is the actual contract every dictionary must satisfy.
 */
type Widen<T> = T extends string
  ? string
  : T extends number
    ? number
    : T extends boolean
      ? boolean
      : T extends ReadonlyArray<infer U>
        ? ReadonlyArray<Widen<U>>
        : T extends object
          ? { [K in keyof T]: Widen<T[K]> }
          : T;

/** `locale` stays a real union so callers can switch on it. */
export type Dictionary = Widen<typeof ru> & { locale: Locale };

const DICTIONARIES: Record<Locale, Dictionary> = {
  ru: ru as Dictionary,
  en,
};

export function getDictionary(locale: Locale): Dictionary {
  return DICTIONARIES[locale] ?? DICTIONARIES[defaultLocale];
}
