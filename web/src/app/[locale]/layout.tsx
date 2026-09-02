import type { Metadata } from "next";
import { notFound } from "next/navigation";
import type { ReactNode } from "react";

import { AppShell } from "@/components/layout/app-shell";
import { LocaleProvider } from "@/components/locale-provider";
import { getDictionary, isLocale, locales, type Locale } from "@/i18n";

export function generateStaticParams() {
  return locales.map((locale) => ({ locale }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>;
}): Promise<Metadata> {
  const { locale } = await params;
  const dict = getDictionary(isLocale(locale) ? locale : "ru");
  return { title: dict.meta.title, description: dict.meta.description };
}

/**
 * Locale layout.
 *
 * This is where the language is decided for the whole tree: the dictionary is
 * resolved on the server and handed to `LocaleProvider` as a plain object, so
 * client components read strings from context instead of importing a
 * dictionary and shipping every locale to the browser.
 */
export default async function LocaleLayout({
  children,
  params,
}: {
  children: ReactNode;
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  if (!isLocale(locale)) notFound();

  const typed = locale as Locale;
  const dict = getDictionary(typed);

  return (
    <LocaleProvider dict={dict} locale={typed}>
      <div lang={typed}>
        <AppShell>{children}</AppShell>
      </div>
    </LocaleProvider>
  );
}
