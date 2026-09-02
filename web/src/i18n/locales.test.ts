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
} from "./locales";

describe("locale set", () => {
  it("puts Russian first and makes it the default", () => {
    expect(locales[0]).toBe("ru");
    expect(defaultLocale).toBe("ru");
  });

  it("names every locale in both maps", () => {
    for (const locale of locales) {
      expect(LOCALE_NAMES[locale]).toBeTruthy();
      expect(LOCALE_SHORT[locale]).toBeTruthy();
    }
  });

  it("remembers the choice for a year", () => {
    expect(LOCALE_COOKIE).toBe("FT12_LOCALE");
    expect(LOCALE_COOKIE_MAX_AGE).toBe(60 * 60 * 24 * 365);
  });
});

describe("isLocale", () => {
  it.each(locales)("accepts %s", (locale) => {
    expect(isLocale(locale)).toBe(true);
  });

  it.each(["de", "", "RU", "ru-RU"])("rejects %s", (value) => {
    expect(isLocale(value)).toBe(false);
  });
});

describe("getLocaleFromPathname", () => {
  it.each([
    { path: "/ru", want: "ru" },
    { path: "/en/monitor", want: "en" },
    { path: "/ru/protocol?x=1", want: "ru" },
    { path: "/", want: defaultLocale },
    { path: "", want: defaultLocale },
    { path: "/monitor", want: defaultLocale },
    { path: "/de/monitor", want: defaultLocale },
  ])("reads $path as $want", ({ path, want }) => {
    expect(getLocaleFromPathname(path)).toBe(want);
  });
});

describe("withLocale", () => {
  it.each([
    { path: "/ru/monitor", locale: "en" as const, want: "/en/monitor" },
    { path: "/en", locale: "ru" as const, want: "/ru" },
    { path: "/", locale: "en" as const, want: "/en" },
    { path: "/monitor", locale: "en" as const, want: "/en/monitor" },
    { path: "/ru/protocol/deep", locale: "en" as const, want: "/en/protocol/deep" },
  ])("rewrites $path to $want", ({ path, locale, want }) => {
    expect(withLocale(path, locale)).toBe(want);
  });

  it("is idempotent for the same locale", () => {
    expect(withLocale("/ru/monitor", "ru")).toBe("/ru/monitor");
  });
});
