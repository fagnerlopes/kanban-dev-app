import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [tailwindcss(), reactRouter()],
  build: {
    // Ship source maps next to the bundles. Without them a browser error in
    // Sentry reads "a.b is not a function" at "index-4f2a.js:1:20481", which
    // is useless to hand to an agent -- with them it points at the real file
    // and line.
    //
    // The maps are served publicly by the Go static handler and Sentry fetches
    // them from the stack-trace URL, so this needs no SENTRY_AUTH_TOKEN and no
    // upload step. The alternative (@sentry/vite-plugin) would need that token
    // at DOCKER BUILD time -- the same build-time/runtime trap that
    // VITE_SENTRY_DSN falls into. See docs/adr/004-sentry-dsn-em-runtime.md.
    //
    // Trade-off: the frontend source becomes readable from the deployed app.
    // Accepted -- this repo is public, and the workshop forks are public too.
    sourcemap: true,
  },
  resolve: {
    tsconfigPaths: true,
  },
  server: {
    host: true, // bind all interfaces (reachable via container bridge IP)
    proxy: {
      "/api": "http://localhost:8080",
      "/auth": "http://localhost:8080",
    },
  },
});
