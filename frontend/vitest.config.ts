import { fileURLToPath } from "node:url";
import { defineConfig } from "vitest/config";

// Standalone Vitest config — deliberately does NOT merge the app's Vite
// config. Merging pulls in the React Router Vite plugin, which expects its
// own preamble and breaks in the test environment. Vitest transforms JSX
// natively, so no React plugin is needed here. The `~` alias mirrors the
// tsconfig paths (~/ -> ./app/).
export default defineConfig({
  resolve: {
    alias: {
      "~": fileURLToPath(new URL("./app", import.meta.url)),
    },
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: [fileURLToPath(new URL("./test/setup.ts", import.meta.url))],
    include: ["app/**/*.test.{ts,tsx}"],
  },
});
