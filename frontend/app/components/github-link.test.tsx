import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { GithubLink } from "./github-link";

const FALLBACK = "https://github.com/fagnerlopes/kanban-dev-app";

beforeEach(() => {
  vi.resetModules();
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("GithubLink", () => {
  it("renderiza um link acessível para o GitHub", async () => {
    vi.stubGlobal("fetch", vi.fn(() => new Promise(() => {})));

    render(<GithubLink />);

    const link = await screen.findByRole("link", { name: "Ver o código no GitHub" });
    expect(link).toHaveAttribute("target", "_blank");
    expect(link.getAttribute("rel")).toContain("noreferrer");
  });

  // O ícone não pode aparecer sem destino enquanto /api/config não responde.
  it("usa o repositório padrão antes da configuração chegar", () => {
    vi.stubGlobal("fetch", vi.fn(() => new Promise(() => {})));

    render(<GithubLink />);

    expect(screen.getByRole("link")).toHaveAttribute("href", FALLBACK);
  });

  // Cada participante forka: o link tem de apontar para o fork dele, não para
  // o repositório de origem.
  it("passa a apontar para o repositório devolvido por /api/config", async () => {
    const forked = "https://github.com/participante/kanban-dev-app";
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => ({ ok: true, json: async () => ({ repo_url: forked }) })),
    );

    const { GithubLink: Fresh } = await import("./github-link");
    render(<Fresh />);

    await waitFor(() =>
      expect(screen.getByRole("link")).toHaveAttribute("href", forked),
    );
  });

  // Se /api/config falhar, o cabeçalho não pode quebrar junto.
  it("mantém o link quando a configuração não carrega", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: false })));

    const { GithubLink: Fresh } = await import("./github-link");
    render(<Fresh />);

    await waitFor(() =>
      expect(screen.getByRole("link")).toHaveAttribute("href", FALLBACK),
    );
  });
});
