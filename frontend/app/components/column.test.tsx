import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Column, columnAccent } from "./column";
import type { ColumnWithTasks } from "~/lib/api";

describe("columnAccent", () => {
  it("maps each known column to its own tint token", () => {
    expect(columnAccent("Backlog")).toBe("var(--col-backlog)");
    expect(columnAccent("To Do")).toBe("var(--col-todo)");
    expect(columnAccent("In Dev")).toBe("var(--col-indev)");
    expect(columnAccent("Review")).toBe("var(--col-review)");
    expect(columnAccent("Done")).toBe("var(--col-done)");
  });

  it("falls back to the accent for unknown columns", () => {
    expect(columnAccent("Something Else")).toBe("var(--color-accent)");
  });
});

const column: ColumnWithTasks = {
  id: 2,
  name: "To Do",
  position: 1,
  tasks: [],
};

describe("Column", () => {
  it("shows the column name and an empty state", () => {
    render(
      <Column
        column={column}
        onDragStart={() => {}}
        onDropTask={() => {}}
        onDelete={() => {}}
        onAdd={() => {}}
      />,
    );
    expect(screen.getByText("To Do")).toBeInTheDocument();
    expect(screen.getByText("Vazio")).toBeInTheDocument();
  });

  it("shows the task count in the header badge", () => {
    const withTask: ColumnWithTasks = {
      ...column,
      tasks: [{ id: 1, column_id: 2, title: "T1", description: "", position: 0 }],
    };
    render(
      <Column
        column={withTask}
        onDragStart={() => {}}
        onDropTask={() => {}}
        onDelete={() => {}}
        onAdd={() => {}}
      />,
    );
    expect(screen.getByText("1")).toBeInTheDocument();
    expect(screen.getByText("T1")).toBeInTheDocument();
  });

  it("creates a task via the inline form (Enter or button)", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(
      <Column
        column={column}
        onDragStart={() => {}}
        onDropTask={() => {}}
        onDelete={() => {}}
        onAdd={onAdd}
      />,
    );
    // Open the add form.
    await user.click(screen.getByRole("button", { name: "Adicionar em To Do" }));
    const input = screen.getByTestId("new-task-input");
    await user.type(input, "Nova tarefa");
    await user.click(screen.getByRole("button", { name: "Adicionar" }));
    expect(onAdd).toHaveBeenCalledWith(2, "Nova tarefa");
  });

  it("does nothing when the title is blank", async () => {
    const user = userEvent.setup();
    const onAdd = vi.fn();
    render(
      <Column
        column={column}
        onDragStart={() => {}}
        onDropTask={() => {}}
        onDelete={() => {}}
        onAdd={onAdd}
      />,
    );
    await user.click(screen.getByRole("button", { name: "Adicionar em To Do" }));
    const input = screen.getByTestId("new-task-input");
    await user.type(input, "   ");
    await user.click(screen.getByRole("button", { name: "Adicionar" }));
    expect(onAdd).not.toHaveBeenCalled();
  });
});
