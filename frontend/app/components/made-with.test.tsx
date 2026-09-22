import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MadeWith } from "./made-with";

describe("MadeWith", () => {
  it("credits both Cofounder and Locaweb Cloud", () => {
    render(<MadeWith />);

    expect(screen.getByRole("link", { name: "Cofounder" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Locaweb Cloud" })).toBeInTheDocument();
  });

  // target="_blank" without rel="noreferrer" lets the opened page reach back
  // through window.opener — and the credit is the only outbound link in the app.
  it("opens the credits in a new tab without leaking the opener", () => {
    render(<MadeWith />);

    for (const link of screen.getAllByRole("link")) {
      expect(link).toHaveAttribute("target", "_blank");
      expect(link.getAttribute("rel")).toContain("noreferrer");
      expect(link.getAttribute("href")).toMatch(/^https:\/\//);
    }
  });

  // The "//" is decoration, not content: a screen reader announcing "slash
  // slash feito com" would be noise.
  it("hides the comment marker from assistive tech", () => {
    const { container } = render(<MadeWith />);

    const marker = container.querySelector('[aria-hidden="true"]');
    expect(marker).not.toBeNull();
    expect(marker).toHaveTextContent("//");
  });
});
