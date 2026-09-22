import type { ReactNode } from "react";

/**
 * Workshop credit line. Lives in one component so the login screen and the
 * board never drift apart.
 *
 * Styled as a code comment ("// ...") because the whole app is JetBrains Mono
 * on a dotted workbench grid — a plain centered footer would read as bolted
 * on, this reads as part of the surface.
 */
export function MadeWith({ className = "" }: { className?: string }) {
  return (
    <p className={`font-display text-xs text-[var(--color-faint)] ${className}`}>
      <span aria-hidden="true" className="text-[var(--color-accent)]/60">
        {"// "}
      </span>
      Feito com{" "}
      <CreditLink href="https://github.com/locaweb/cofounder">Cofounder</CreditLink>
      {" e "}
      <CreditLink href="https://www.locaweb.com.br/locaweb-cloud/">
        Locaweb Cloud
      </CreditLink>
    </p>
  );
}

// rel="noreferrer" also implies noopener: without it the opened page can reach
// back through window.opener.
//
// The link sits one step brighter than the line around it. --color-faint is
// about 2.6:1 on the dark background -- fine for a decorative line, not for
// something clickable; --color-muted clears 7:1 in both themes and doubles as
// the hint that these words are links.
function CreditLink({ href, children }: { href: string; children: ReactNode }) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noreferrer"
      className="cursor-pointer rounded-sm text-[var(--color-muted)] underline decoration-dotted underline-offset-2 transition-colors hover:text-[var(--color-accent)] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--color-accent)]"
    >
      {children}
    </a>
  );
}
