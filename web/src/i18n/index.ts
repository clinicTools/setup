import { createI18n } from "vue-i18n";
import de from "./de.json";
import en from "./en.json";

export type Locale = "de" | "en";

const STORAGE_KEY = "da-locale";

function detectLocale(): Locale {
  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored === "de" || stored === "en") return stored;
  return navigator.language.startsWith("en") ? "en" : "de";
}

// Deutsch ist die Domänensprache (Fallback); Englisch als Alternative.
export const i18n = createI18n({
  legacy: false,
  locale: detectLocale(),
  fallbackLocale: "de",
  messages: { de, en },
});

/** Setzt und persistiert die aktive Sprache. */
export function setLocale(locale: Locale): void {
  i18n.global.locale.value = locale;
  localStorage.setItem(STORAGE_KEY, locale);
  document.documentElement.lang = locale;
}
