import {
  isRouteErrorResponse,
  Links,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
} from "react-router";

import type { Route } from "./+types/root";
import "./app.css";

// Inline theme script: owns the data-theme attribute and runs before paint so
// there is no theme flash. Persists to localStorage; defaults to "system"
// (follows the OS preference). React never renders this attribute.
const themeScript = `(function(){try{var t=localStorage.getItem("kbd-theme")||"system";var d=t==="dark"||((t==="system")&&window.matchMedia("(prefers-color-scheme: dark)").matches);document.documentElement.setAttribute("data-theme",d?"dark":"light");}catch(e){document.documentElement.setAttribute("data-theme","light");}})();`;

// Favicon: a kanban column glyph — three stacked bars, terminal-green.
const favicon =
  "data:image/svg+xml," +
  "<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16'>" +
  "<rect x='1' y='2' width='3.4' height='12' rx='1' fill='%2316a34a'/>" +
  "<rect x='6.3' y='2' width='3.4' height='8' rx='1' fill='%234ade80'/>" +
  "<rect x='11.6' y='2' width='3.4' height='5' rx='1' fill='%23fbbf24'/>" +
  "</svg>";

export const links: Route.LinksFunction = () => [
  { rel: "icon", href: favicon },
  { rel: "preconnect", href: "https://fonts.googleapis.com" },
  { rel: "preconnect", href: "https://fonts.gstatic.com", crossOrigin: "anonymous" },
  {
    rel: "stylesheet",
    href: "https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600;700&display=swap",
  },
];

export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="pt-BR" suppressHydrationWarning>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>Kanban Dev Flow</title>
        <script dangerouslySetInnerHTML={{ __html: themeScript }} />
        <Meta />
        <Links />
      </head>
      <body>
        {children}
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  );
}

export default function App() {
  return <Outlet />;
}

export function ErrorBoundary({ error }: Route.ErrorBoundaryProps) {
  let message = "Ops!";
  let details = "Um erro inesperado ocorreu.";
  let stack: string | undefined;

  if (isRouteErrorResponse(error)) {
    message = error.status === 404 ? "404" : "Erro";
    details =
      error.status === 404
        ? "A página solicitada não foi encontrada."
        : error.statusText || details;
  } else if (import.meta.env.DEV && error && error instanceof Error) {
    details = error.message;
    stack = error.stack;
  }

  return (
    <main className="min-h-screen flex flex-col items-center justify-center gap-3 p-4 text-center">
      <h1 className="text-3xl font-display font-bold text-[var(--color-fg)]">
        {message}
      </h1>
      <p className="text-[var(--color-muted)]">{details}</p>
      {stack && (
        <pre className="w-full max-w-2xl p-4 overflow-x-auto rounded-lg text-left text-sm bg-[var(--color-surface)]">
          <code>{stack}</code>
        </pre>
      )}
    </main>
  );
}
