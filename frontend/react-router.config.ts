import type { Config } from "@react-router/dev/config";

export default {
  // Pure SPA at runtime — no Node server, ever.
  ssr: false,
} satisfies Config;
