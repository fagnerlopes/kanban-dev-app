import { useEffect, useState } from "react";

export type AppConfig = {
  sentry_dsn: string;
  environment: string;
  release: string;
  repo_url: string;
};

// Fallback usado até /api/config responder, para nada renderizar vazio.
const FALLBACK: AppConfig = {
  sentry_dsn: "",
  environment: "local",
  release: "kanban-dev-app",
  repo_url: "https://github.com/fagnerlopes/kanban-dev-app",
};

let cached: AppConfig | null = null;

export function useAppConfig(): AppConfig {
  const [config, setConfig] = useState<AppConfig>(cached ?? FALLBACK);

  useEffect(() => {
    if (cached) return;
    let active = true;
    fetch("/api/config")
      .then((r) => (r.ok ? r.json() : null))
      .then((data: AppConfig | null) => {
        if (!data || !active) return;
        cached = { ...FALLBACK, ...data };
        setConfig(cached);
      })
      .catch(() => {});
    return () => {
      active = false;
    };
  }, []);

  return config;
}
