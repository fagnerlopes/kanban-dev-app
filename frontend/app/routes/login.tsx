import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import { KanbanSquare, Loader2, Moon, Sun } from "lucide-react";
import { MadeWith } from "~/components/made-with";
import { Button } from "~/components/ui/button";
import { useAuth } from "~/hooks/use-auth";
import { useTheme } from "~/hooks/use-theme";

export default function Login() {
  const { session, login } = useAuth();
  const { theme, toggle } = useTheme();
  const navigate = useNavigate();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (session) navigate("/board", { replace: true });
  }, [session, navigate]);

  async function onLogin() {
    setBusy(true);
    setError(null);
    try {
      await login();
      navigate("/board", { replace: true });
    } catch (e) {
      setError(e instanceof Error ? e.message : "Não foi possível entrar.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="min-h-screen bg-workbench flex flex-col">
      <header className="flex items-center justify-between px-6 py-4">
        <div className="flex items-center gap-2 font-display font-semibold tracking-tight">
          <KanbanSquare className="size-5 text-[var(--color-accent)]" />
          <span className="text-[var(--color-fg)]">Kanban Dev Flow</span>
        </div>
        <button
          onClick={toggle}
          aria-label="Alternar tema"
          className="cursor-pointer p-2 rounded-lg text-[var(--color-muted)] hover:bg-[var(--color-surface)] transition-colors"
        >
          {theme === "dark" ? <Moon className="size-4" /> : <Sun className="size-4" />}
        </button>
      </header>

      <div className="flex-1 flex items-center justify-center p-4">
        <div className="w-full max-w-sm rounded-2xl border border-[var(--color-border)] bg-[var(--color-bg-elevated)] p-8 shadow-[var(--shadow-card)]">
          <div className="mb-6 flex flex-col items-center gap-2 text-center">
            <div className="flex size-12 items-center justify-center rounded-xl bg-[var(--color-accent-soft)]">
              <KanbanSquare className="size-6 text-[var(--color-accent)]" />
            </div>
            <h1 className="font-display text-xl font-semibold text-[var(--color-fg)]">
              Bem-vindo de volta
            </h1>
            <p className="text-sm text-[var(--color-muted)]">
              Entre para acessar o seu quadro.
            </p>
          </div>

          <Button
            onClick={onLogin}
            disabled={busy}
            className="w-full h-10"
            data-testid="login-button"
          >
            {busy ? (
              <>
                <Loader2 className="size-4 animate-spin" />
                Entrando…
              </>
            ) : (
              "Entrar como demo"
            )}
          </Button>

          {error && (
            <p className="mt-3 text-center text-sm text-[var(--color-destructive)]">
              {error}
            </p>
          )}

          <p className="mt-6 text-center text-xs text-[var(--color-faint)]">
            Demo de workshop — usuário único, sem senha.
          </p>
        </div>
      </div>

      <footer className="px-6 pb-6 text-center">
        <MadeWith />
      </footer>
    </main>
  );
}
