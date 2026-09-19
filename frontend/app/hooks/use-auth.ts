import { useCallback, useState } from "react";

// Mock auth: a single demo user, as decided in docs/adr/002-mock-login.md.
// The session lives entirely in localStorage — there is no server-side session
// and no per-user data, so the "login" is a client-side gate, not a security
// boundary.
//
// It deliberately does NOT call POST /api/dev/login: that route only exists
// when DEV_MODE=1, which is a local-only flag. Depending on it here would mean
// either a broken login in every deployed environment, or setting DEV_MODE in
// production — which also stops the Go server from serving this very SPA.
const TOKEN_KEY = "kbd-token";
const USER_KEY = "kbd-user";

const DEMO_USER = "demo@kanban.local";

export interface Session {
  user: string;
  token: string;
}

function readSession(): Session | null {
  try {
    const token = localStorage.getItem(TOKEN_KEY);
    const user = localStorage.getItem(USER_KEY);
    return token && user ? { user, token } : null;
  } catch {
    return null; // private mode / blocked site data
  }
}

export function useAuth() {
  const [session, setSession] = useState<Session | null>(readSession);

  const login = useCallback(async (): Promise<Session> => {
    const next: Session = { user: DEMO_USER, token: crypto.randomUUID() };
    try {
      localStorage.setItem(TOKEN_KEY, next.token);
      localStorage.setItem(USER_KEY, next.user);
    } catch {
      // Session stays in memory for this tab only.
    }
    setSession(next);
    return next;
  }, []);

  const logout = useCallback(() => {
    try {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);
    } catch {
      // Nothing persisted to clear.
    }
    setSession(null);
  }, []);

  return { session, isAuthenticated: !!session, login, logout };
}
