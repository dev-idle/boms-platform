import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTs from "eslint-config-next/typescript";

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTs,
  globalIgnores([".next/**", "out/**", "build/**", "next-env.d.ts"]),
  {
    name: "boms/dal-boundary",
    files: ["src/**/*.{ts,tsx}"],
    ignores: ["src/lib/api-client.ts", "src/lib/dal/**/*"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          paths: [
            {
              name: "@/lib/api-client",
              message:
                "Backend access must go through @/lib/dal (server-only boundary with Zod validation).",
            },
          ],
        },
      ],
    },
  },
]);

export default eslintConfig;
