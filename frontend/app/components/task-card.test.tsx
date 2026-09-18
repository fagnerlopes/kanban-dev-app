import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TaskCard } from "./task-card";
import type { Task } from "~/lib/api";

const task: Task = {
  id: 1,
  column_id: 1,
  title: "Corrigir bug no login",
  description: "500 na arrancada",
  position: 0,
};

describe("TaskCard", () => {
  it("renders the task title", () => {
    render(
      <TaskCard
        task={task}
        accent="#16a34a"
        onDragStart={() => {}}
        onDelete={() => {}}
      />,
    );
    expect(screen.getByText("Corrigir bug no login")).toBeInTheDocument();
  });

  it("renders the description when present", () => {
    render(
      <TaskCard
        task={task}
        accent="#16a34a"
        onDragStart={() => {}}
        onDelete={() => {}}
      />,
    );
    expect(screen.getByText("500 na arrancada")).toBeInTheDocument();
  });

  it("calls onDelete when the delete button is clicked", async () => {
    const user = userEvent.setup();
    const onDelete = vi.fn();
    render(
      <TaskCard task={task} accent="#16a34a" onDragStart={() => {}} onDelete={onDelete} />,
    );
    await user.click(screen.getByRole("button", { name: "Remover Corrigir bug no login" }));
    expect(onDelete).toHaveBeenCalledWith(task);
  });

  it("sets the per-column accent on the left border", () => {
    render(
      <TaskCard
        task={task}
        accent="rgb(14, 165, 233)"
        onDragStart={() => {}}
        onDelete={() => {}}
      />,
    );
    const card = screen.getByTestId("task-card");
    expect(card.style.borderLeft).toContain("rgb(14, 165, 233)");
  });
});
