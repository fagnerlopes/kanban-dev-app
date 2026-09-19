import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router";
import { KanbanSquare, LogOut, Moon, Sun } from "lucide-react";
import { Column } from "~/components/column";
import { useAuth } from "~/hooks/use-auth";
import { useTheme } from "~/hooks/use-theme";
import {
  createTask,
  deleteTask,
  fetchBoard,
  updateTask,
  type Board,
} from "~/lib/api";

export default function Board() {
  const { session, isAuthenticated, logout } = useAuth();
  const { theme, toggle } = useTheme();
  const navigate = useNavigate();
  const [board, setBoard] = useState<Board | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [draggingId, setDraggingId] = useState<number | null>(null);

  // Guard: not logged in -> back to login.
  useEffect(() => {
    if (!isAuthenticated) navigate("/", { replace: true });
  }, [isAuthenticated, navigate]);

  const load = useCallback(async () => {
    try {
      setError(null);
      setBoard(await fetchBoard());
    } catch (e) {
      setError(e instanceof Error ? e.message : "Erro ao carregar o quadro.");
    }
  }, []);

  useEffect(() => {
    if (isAuthenticated) load();
  }, [isAuthenticated, load]);

  function handleDragStart(e: React.DragEvent, task: { id: number }) {
    e.dataTransfer.setData("text/task-id", String(task.id));
    e.dataTransfer.effectAllowed = "move";
    setDraggingId(task.id);
  }

  async function handleDrop(columnId: number, taskId: number) {
    if (draggingId === null) return;
    setDraggingId(null);
    try {
      await updateTask(taskId, {
        column_id: columnId,
        position: 0,
        title: (board?.columns.flatMap((c) => c.tasks) ?? []).find(
          (t) => t.id === taskId,
        )?.title ?? "",
      });
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Erro ao mover tarefa.");
    }
  }

  async function handleAdd(columnId: number, title: string) {
    try {
      await createTask({ column_id: columnId, title });
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Erro ao criar tarefa.");
    }
  }

  async function handleDelete(task: { id: number; title: string }) {
    try {
      await deleteTask(task.id);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Erro ao remover tarefa.");
    }
  }

  function handleLogout() {
    logout();
    navigate("/", { replace: true });
  }

  return (
    <main className="h-screen bg-workbench flex flex-col overflow-hidden">
      {/* Top bar */}
      <header className="flex items-center justify-between border-b border-[var(--color-border)] bg-[var(--color-bg-elevated)]/80 backdrop-blur px-6 py-3">
        <div className="flex items-center gap-2 font-display font-semibold tracking-tight">
          <KanbanSquare className="size-5 text-[var(--color-accent)]" />
          <span className="text-[var(--color-fg)]">Kanban Dev Flow</span>
        </div>
        <div className="flex items-center gap-1.5">
          <button
            onClick={toggle}
            aria-label="Alternar tema"
            className="cursor-pointer p-2 rounded-lg text-[var(--color-muted)] hover:bg-[var(--color-surface)] transition-colors"
          >
            {theme === "dark" ? <Moon className="size-4" /> : <Sun className="size-4" />}
          </button>
          <div className="flex items-center gap-2 rounded-lg bg-[var(--color-surface)] px-2.5 py-1.5">
            <span className="size-2 rounded-full bg-[var(--color-accent)]" />
            <span className="text-xs text-[var(--color-muted)]">{session?.user}</span>
          </div>
          <button
            onClick={handleLogout}
            aria-label="Sair"
            className="cursor-pointer p-2 rounded-lg text-[var(--color-muted)] hover:bg-[var(--color-surface)] transition-colors"
          >
            <LogOut className="size-4" />
          </button>
        </div>
      </header>

      {error && (
        <div className="mx-6 mt-4 rounded-lg border border-[var(--color-destructive)]/30 bg-[var(--color-destructive)]/10 px-4 py-2 text-sm text-[var(--color-destructive)]">
          {error}
        </div>
      )}

      {/* Board */}
      <div className="flex-1 overflow-x-auto overflow-y-hidden p-4">
        <div className="flex h-full gap-4">
          {board?.columns.map((col) => (
            <Column
              key={col.id}
              column={col}
              onDragStart={handleDragStart}
              onDropTask={handleDrop}
              onDelete={handleDelete}
              onAdd={handleAdd}
            />
          ))}
        </div>
      </div>
    </main>
  );
}
