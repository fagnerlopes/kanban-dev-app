import { useCallback, useState } from "react";

// Mock auth: a single demo user. The "token" is a mock session string from
// POST /api/dev/login (DEV_MODE only). No real auth — this is a workshop demo.
const TOKEN_KEY = "kbd-token";
const USER_KEY = "kbd-user";

export interface Session {
  user: string;
  token: string;
}

export function useAuth() {
  const [session, setSession] = useState<Session | null>(() => {
    const token = localStorage.getItem(TOKEN_KEY);
    const user = localStorage.getItem(USER_KEY);
    return token && user ? { user, token } : null;
  });

  const login = useCallback(async () => {
    const res = await fetch("/api/dev/login", { method: "POST" });
    if (!res.ok) throw new Error("login failed: " + res.status);
    const data = (await res.json()) as { user: string; token: string };
    localStorage.setItem(TOKEN_KEY, data.token);
    localStorage.setItem(USER_KEY, data.user);
    setSession({ user: data.user, token: data.token });
    return data;
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    setSession(null);
  }, []);

  return { session, isAuthenticated: !!session, login, logout };
}
