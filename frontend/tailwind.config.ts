import type { Config } from "tailwindcss";

/**
 * Tailwind v4: design tokens live in `src/app/globals.css` (`@theme inline`).
 * This file only declares content paths for tooling (shadcn, ESLint).
 */
export default {
  content: ["./src/**/*.{ts,tsx}"],
} satisfies Config;
