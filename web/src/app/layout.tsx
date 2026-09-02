import type { Metadata } from "next";
import type { ReactNode } from "react";

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
    <html suppressHydrationWarning>
      <body className="min-h-screen antialiased">{children}</body>
    </html>
  );
}
