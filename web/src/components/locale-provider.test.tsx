import { render, screen } from "@testing-library/react";

import { getDictionary } from "@/i18n";

import { LocaleProvider, useDictionary, useLocale } from "./locale-provider";

function Consumer() {
  const dict = useDictionary();
  const locale = useLocale();
  return (
    <span data-testid="value">
      {locale}:{dict.nav.overview}
    </span>
  );
}

describe("LocaleProvider", () => {
  it.each([
    { locale: "ru" as const, want: "ru:Обзор" },
    { locale: "en" as const, want: "en:Overview" },
  ])("hands $locale strings to consumers", ({ locale, want }) => {
    render(
      <LocaleProvider dict={getDictionary(locale)} locale={locale}>
        <Consumer />
      </LocaleProvider>,
    );

    expect(screen.getByTestId("value")).toHaveTextContent(want);
  });
});

describe("useDictionary outside a provider", () => {
  it("fails loudly rather than rendering untranslated", () => {
    // The thrown error is the point; React also logs it, which is noise here.
    const consoleError = jest.spyOn(console, "error").mockImplementation(() => {});

    expect(() => render(<Consumer />)).toThrow(/useDictionary must be used inside/);

    consoleError.mockRestore();
  });
});

describe("useLocale outside a provider", () => {
  it("fails loudly as well", () => {
    const consoleError = jest.spyOn(console, "error").mockImplementation(() => {});

    function LocaleOnly() {
      return <span>{useLocale()}</span>;
    }

    expect(() => render(<LocaleOnly />)).toThrow(/useLocale must be used inside/);

    consoleError.mockRestore();
  });
});
