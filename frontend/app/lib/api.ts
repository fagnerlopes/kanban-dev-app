// Thin fetch wrapper for the Kanban Dev Flow API.
export interface Column {
  id: number;
  name: string;
  position: number;
}
export interface Task {
  id: number;
  column_id: number;
  title: string;
  description: string;
  position: number;
}
export interface ColumnWithTasks extends Column {
  tasks: Task[];
}
export interface Board {
  columns: ColumnWithTasks[];
}

export async function fetchBoard(): Promise<Board> {
  const res = await fetch("/api/board");
  if (!res.ok) throw new Error("board: " + res.status);
  return (await res.json()) as Board;
}

export async function createTask(input: {
  column_id: number;
  title: string;
  description?: string;
}): Promise<Task> {
  const res = await fetch("/api/tasks", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!res.ok) throw new Error("createTask: " + res.status);
  return (await res.json()) as Task;
}

export async function updateTask(
  id: number,
  input: { column_id: number; position: number; title: string },
): Promise<Task> {
  const res = await fetch(`/api/tasks/${id}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!res.ok) throw new Error("updateTask: " + res.status);
  return (await res.json()) as Task;
}

export async function deleteTask(id: number): Promise<void> {
  const res = await fetch(`/api/tasks/${id}`, { method: "DELETE" });
  if (!res.ok) throw new Error("deleteTask: " + res.status);
}
