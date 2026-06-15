import type { Config } from "tailwindcss";

/**
 * Tailwind v4: design tokens live in `src/styles/` (imported via `src/app/globals.css`).
 * This file only declares content paths for tooling (shadcn, ESLint).
 */
export default {
  content: ["./src/**/*.{ts,tsx}"],
} satisfies Config;
