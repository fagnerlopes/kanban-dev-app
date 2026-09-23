import type { ReactNode } from "react";

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
