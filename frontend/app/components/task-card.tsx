import { GripVertical, Trash2 } from "lucide-react";
import type { Task } from "~/lib/api";

interface TaskCardProps {
  task: Task;
  accent: string; // CSS color for the left bar (per-column tint)
  onDragStart: (e: React.DragEvent, task: Task) => void;
  onDelete: (task: Task) => void;
}

export function TaskCard({ task, accent, onDragStart, onDelete }: TaskCardProps) {
  return (
    <div
      draggable
      onDragStart={(e) => onDragStart(e, task)}
      data-testid="task-card"
      data-task-id={task.id}
      className="group cursor-grab active:cursor-grabbing rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-elevated)] p-3 shadow-[var(--shadow-card)] transition-shadow hover:shadow-[var(--shadow-card-hover)]"
      style={{ borderLeft: `3px solid ${accent}` }}
    >
      <div className="flex items-start gap-2">
        <GripVertical className="mt-0.5 size-4 shrink-0 text-[var(--color-faint)] opacity-0 group-hover:opacity-100 transition-opacity" />
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium leading-snug text-[var(--color-fg)] break-words">
            {task.title}
          </p>
          {task.description && (
            <p className="mt-1 text-xs text-[var(--color-muted)] line-clamp-2">
              {task.description}
            </p>
          )}
        </div>
        <button
          onClick={() => onDelete(task)}
          aria-label={`Remover ${task.title}`}
          className="cursor-pointer rounded p-1 text-[var(--color-faint)] opacity-0 group-hover:opacity-100 hover:bg-[var(--color-surface)] hover:text-[var(--color-destructive)] transition-all"
        >
          <Trash2 className="size-3.5" />
        </button>
      </div>
    </div>
  );
}
