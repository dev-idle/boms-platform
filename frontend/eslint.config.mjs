import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTs from "eslint-config-next/typescript";

const DAL_BOUNDARY = {
  name: "@/lib/api-client",
  message: "Backend access must go through @/lib/dal (server-only boundary with Zod validation).",
};

const SLICE_BARREL_ONLY = {
  group: ["@/features/*/**", "!@/features/*/server"],
  message:
    "Import a feature slice through its public barrel (@/features/<slice>); inside a slice use relative paths. Only @/features/<slice>/server is a second entry point.",
};

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTs,
  globalIgnores([".next/**", "out/**", "build/**", "next-env.d.ts"]),
  {
    name: "boms/react-settings",
    settings: {
      react: { version: "19" },
    },
  },
  // Flat config does not merge options of the same rule across objects, so each
  // file group states every boundary that applies to it.
  {
    name: "boms/import-boundaries",
    files: ["src/**/*.{ts,tsx}"],
    ignores: ["src/lib/**"],
    rules: {
      "no-restricted-imports": ["error", { paths: [DAL_BOUNDARY], patterns: [SLICE_BARREL_ONLY] }],
    },
  },
  {
    name: "boms/lib-boundaries",
    files: ["src/lib/**/*.{ts,tsx}"],
    ignores: ["src/lib/api-client.ts", "src/lib/dal/**"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          paths: [DAL_BOUNDARY],
          patterns: [
            {
              group: ["@/features/*"],
              message: "lib/ must not import features/* — move shared contracts to lib/.",
            },
          ],
        },
      ],
    },
  },
]);

export default eslintConfig;
