import { redirect } from "next/navigation";

import { defaultLocale } from "@/i18n";

/** The bare root has no language of its own — hand it to the default locale. */
export default function RootPage() {
  redirect(`/${defaultLocale}`);
}
