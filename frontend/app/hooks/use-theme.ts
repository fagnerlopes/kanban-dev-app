import { useCallback, useSyncExternalStore } from "react";

type Theme = "light" | "dark" | "system";
const KEY = "kbd-theme";

function readTheme(): Theme {
  const v = localStorage.getItem(KEY);
  if (v === "light" || v === "dark" || v === "system") return v;
  return "light";
}

function applyTheme(theme: Theme) {
  const dark =
    theme === "dark" ||
    (theme === "system" &&
      window.matchMedia("(prefers-color-scheme: dark)").matches);
  document.documentElement.setAttribute("data-theme", dark ? "dark" : "light");
}

let listeners = new Set<() => void>();
function subscribe(cb: () => void) {
  listeners.add(cb);
  return () => {
    listeners.delete(cb);
  };
}
function emit() {
  listeners.forEach((l) => l());
}

function setThemeLocal(theme: Theme) {
  localStorage.setItem(KEY, theme);
  applyTheme(theme);
  emit();
}

applyTheme(readTheme());
window
  .matchMedia("(prefers-color-scheme: dark)")
  .addEventListener("change", () => {
    if (readTheme() === "system") {
      applyTheme("system");
      emit();
    }
  });

export function useTheme() {
  const theme = useSyncExternalStore(subscribe, readTheme);
  const setTheme = useCallback((t: Theme) => setThemeLocal(t), []);
  const toggle = useCallback(() => {
    const current = readTheme();
    const next: Theme = current === "light" ? "dark" : "light";
    setThemeLocal(next);
  }, []);
  return { theme, setTheme, toggle };
}
