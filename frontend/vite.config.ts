import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [tailwindcss(), reactRouter()],
  build: {
    // O Sentry busca os .map pela URL do bundle; sem eles a stack trace do
    // navegador chega minificada.
    sourcemap: true,
  },
  resolve: {
    tsconfigPaths: true,
  },
  server: {
    host: true,
    proxy: {
      "/api": "http://localhost:8080",
      "/auth": "http://localhost:8080",
    },
  },
});
