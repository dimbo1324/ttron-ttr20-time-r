import Link from "next/link";

import { defaultLocale, getDictionary } from "@/i18n";

export default function NotFound() {
  const dict = getDictionary(defaultLocale);
  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-4 px-6 text-center">
      <p className="font-mono text-5xl font-semibold text-primary">404</p>
      <p className="text-muted-foreground">{dict.meta.title}</p>
      <Link
        href={`/${defaultLocale}`}
        className="rounded-md border border-border-strong px-3 py-1.5 text-sm hover:border-primary/50"
      >
        {dict.nav.overview}
      </Link>
    </main>
  );
}
