import { getDictionary, locales, type Dictionary } from "./index";

/** Every leaf path in an object, so two dictionaries can be compared by shape. */
function keyPaths(value: unknown, prefix = ""): string[] {
  if (typeof value !== "object" || value === null) return [prefix];
  return Object.entries(value as Record<string, unknown>)
    .flatMap(([key, child]) => keyPaths(child, prefix ? `${prefix}.${key}` : key))
    .sort();
}

describe("getDictionary", () => {
  it.each(locales)("returns the %s dictionary", (locale) => {
    expect(getDictionary(locale).locale).toBe(locale);
  });

  it("falls back to the default for an unknown locale", () => {
    expect(getDictionary("de" as never).locale).toBe("ru");
  });
});

describe("dictionary parity", () => {
  const reference = getDictionary("ru");

  it.each(locales.filter((locale) => locale !== "ru"))(
    "%s carries exactly the keys Russian does",
    (locale) => {
      expect(keyPaths(getDictionary(locale))).toEqual(keyPaths(reference));
    },
  );

  it.each(locales)("%s leaves no string empty", (locale) => {
    const empties: string[] = [];

    const walk = (value: unknown, path: string) => {
      if (typeof value === "string") {
        if (value.trim() === "") empties.push(path);
        return;
      }
      if (typeof value !== "object" || value === null) return;
      for (const [key, child] of Object.entries(value as Record<string, unknown>)) {
        walk(child, path ? `${path}.${key}` : key);
      }
    };

    walk(getDictionary(locale), "");
    expect(empties).toEqual([]);
  });

  it("keeps the sections the UI navigates by", () => {
    const dict: Dictionary = getDictionary("ru");

    for (const key of ["overview", "monitor", "protocol", "emulator", "gateway", "reference"]) {
      expect(dict.nav[key as keyof typeof dict.nav]).toBeTruthy();
    }
  });

  it("covers every frame decode error the analyzer can raise", () => {
    const dict = getDictionary("ru");

    for (const code of [
      "tooShort",
      "invalidStart",
      "invalidStartRepeat",
      "invalidLength",
      "invalidChecksum",
      "invalidEnd",
      "emptyPayload",
      "invalidTimestamp",
      "invalidIdentity",
    ]) {
      expect(dict.protocol.errors[code as keyof typeof dict.protocol.errors]).toBeTruthy();
    }
  });
});
