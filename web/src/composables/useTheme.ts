import { ref, watch } from "vue";

type Theme = "light" | "dark";

const STORAGE_KEY = "da-theme";

function initialTheme(): Theme {
  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored === "light" || stored === "dark") return stored;
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

const theme = ref<Theme>(initialTheme());

function apply(t: Theme): void {
  document.documentElement.classList.toggle("dark", t === "dark");
}

apply(theme.value);
watch(theme, (t) => {
  apply(t);
  localStorage.setItem(STORAGE_KEY, t);
});

/** Globales Hell/Dunkel-Theme (an Windows-Designwechsel angelehnt). */
export function useTheme() {
  function toggle(): void {
    theme.value = theme.value === "dark" ? "light" : "dark";
  }
  return { theme, toggle };
}
