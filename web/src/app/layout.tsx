import type { Metadata } from "next";
import type { ReactNode } from "react";

import { defaultLocale } from "@/i18n";

import "./globals.css";

export const metadata: Metadata = {
  title: "FT1.2 / TTR20",
  description: "Protocol bench",
};

/**
 * Root layout.
 *
 * Deliberately thin: the locale lives one segment down, so the `<html lang>`
 * that actually matters is set by `[locale]/layout.tsx`. This one exists only
 * to own the stylesheet and the document shell.
 */
export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    // `lang` is the default rather than the truth: the locale is a segment
    // below this layout, so `[locale]/layout.tsx` narrows it for its subtree.
    <html lang={defaultLocale} suppressHydrationWarning>
      <body className="min-h-screen antialiased">{children}</body>
    </html>
  );
}
