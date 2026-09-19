import { useState } from "react";
import { Plus } from "lucide-react";
import type { ColumnWithTasks } from "~/lib/api";
import { TaskCard } from "./task-card";
import { Button } from "~/components/ui/button";

// Maps a column name to its CSS accent token.
export function columnAccent(name: string): string {
  switch (name) {
    case "Backlog":
      return "var(--col-backlog)";
    case "To Do":
      return "var(--col-todo)";
    case "In Dev":
      return "var(--col-indev)";
    case "Review":
      return "var(--col-review)";
    case "Done":
      return "var(--col-done)";
    default:
      return "var(--color-accent)";
  }
}

interface ColumnProps {
  column: ColumnWithTasks;
  onDragStart: (e: React.DragEvent, task: { id: number }) => void;
  onDropTask: (columnId: number, taskId: number, index: number) => void;
  onDelete: (task: { id: number; title: string }) => void;
  onAdd: (columnId: number, title: string) => void;
}

export function Column({
  column,
  onDragStart,
  onDropTask,
  onDelete,
  onAdd,
}: ColumnProps) {
  const [isOver, setIsOver] = useState(false);
  const [adding, setAdding] = useState(false);
  const [title, setTitle] = useState("");
  const accent = columnAccent(column.name);

  function submitAdd() {
    const t = title.trim();
    if (t) onAdd(column.id, t);
    setTitle("");
    setAdding(false);
  }

  return (
    <section
      onDragOver={(e) => {
        e.preventDefault();
        setIsOver(true);
      }}
      onDragLeave={() => setIsOver(false)}
      onDrop={(e) => {
        e.preventDefault();
        setIsOver(false);
        const taskId = Number(e.dataTransfer.getData("text/task-id"));
        if (taskId) onDropTask(column.id, taskId, column.tasks.length);
      }}
      data-testid="column"
      data-column-id={column.id}
      className={`flex h-full min-h-[90%] w-72 shrink-0 flex-col rounded-xl border bg-[var(--color-surface)]/50 transition-colors ${
        isOver
          ? "border-[var(--color-accent)] bg-[var(--color-accent-soft)]/40"
          : "border-[var(--color-border)]"
      }`}
    >
      {/* Header: mono, terminal-like */}
      <header className="flex items-center justify-between px-3 py-2.5">
        <div className="flex items-center gap-2">
          <span
            className="size-2.5 rounded-full"
            style={{ backgroundColor: accent }}
          />
          <h2 className="font-mono text-xs font-medium uppercase tracking-wider text-[var(--color-muted)]">
            {column.name}
          </h2>
          <span className="rounded bg-[var(--color-surface)] px-1.5 font-mono text-[11px] text-[var(--color-faint)]">
            {column.tasks.length}
          </span>
        </div>
        <button
          onClick={() => setAdding((v) => !v)}
          aria-label={`Adicionar em ${column.name}`}
          className="cursor-pointer rounded p-1 text-[var(--color-faint)] hover:bg-[var(--color-surface)] hover:text-[var(--color-fg)] transition-colors"
        >
          <Plus className="size-4" />
        </button>
      </header>

      {/* Task list */}
      <div className="flex min-h-0 flex-1 flex-col gap-2 overflow-y-auto px-2 pb-2">
        {adding && (
          <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-elevated)] p-2">
            <input
              autoFocus
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") submitAdd();
                if (e.key === "Escape") setAdding(false);
              }}
              placeholder="Título da tarefa"
              className="w-full bg-transparent text-sm text-[var(--color-fg)] outline-none placeholder:text-[var(--color-faint)]"
              data-testid="new-task-input"
            />
            <div className="mt-2 flex gap-1.5">
              <Button size="sm" onClick={submitAdd}>
                Adicionar
              </Button>
              <Button size="sm" variant="ghost" onClick={() => setAdding(false)}>
                Cancelar
              </Button>
            </div>
          </div>
        )}

        {column.tasks.map((task) => (
          <TaskCard
            key={task.id}
            task={task}
            accent={accent}
            onDragStart={onDragStart}
            onDelete={onDelete}
          />
        ))}

        {column.tasks.length === 0 && !adding && (
          <p className="px-1 py-4 text-center text-xs text-[var(--color-faint)]">
            Vazio
          </p>
        )}
      </div>
    </section>
  );
}
