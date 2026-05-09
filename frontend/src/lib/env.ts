import "server-only";

import { z } from "zod";

const DEV_FALLBACK_SECRET =
  "development-only-boms-internal-proxy-secret-min-32!";

const serverEnvSchema = z.object({
  BOMS_BACKEND_URL: z
    .string()
    .url()
    .describe("Origin of the Go Fiber API (no trailing slash)"),
  INTERNAL_PROXY_SECRET: z
    .string()
    .min(32, "INTERNAL_PROXY_SECRET must be at least 32 characters")
    .describe("Shared secret injected by proxy; must match Fiber trusted header validation"),
});

export type ServerEnv = z.infer<typeof serverEnvSchema>;

let cached: ServerEnv | null = null;

function resolveRawEnv(): Record<string, string | undefined> {
  const isProd = process.env.NODE_ENV === "production";
  return {
    BOMS_BACKEND_URL: process.env.BOMS_BACKEND_URL ?? "http://127.0.0.1:8080",
    INTERNAL_PROXY_SECRET:
      process.env.INTERNAL_PROXY_SECRET ??
      (isProd ? undefined : DEV_FALLBACK_SECRET),
  };
}

/**
 * Validates process-bound configuration once per runtime.
 * Never import this module from Client Components.
 */
export function getServerEnv(): ServerEnv {
  if (cached) {
    return cached;
  }
  const raw = resolveRawEnv();
  const parsed = serverEnvSchema.safeParse(raw);
  if (!parsed.success) {
    const detail = parsed.error.flatten().fieldErrors;
    throw new Error(
      `[BOMS] Invalid server environment: ${JSON.stringify(detail, null, 2)}`,
    );
  }
  if (
    process.env.NODE_ENV !== "production" &&
    raw.INTERNAL_PROXY_SECRET === DEV_FALLBACK_SECRET
  ) {
    console.warn(
      "[BOMS] Using development INTERNAL_PROXY_SECRET fallback. Set INTERNAL_PROXY_SECRET in .env.local before production.",
    );
  }
  cached = parsed.data;
  return parsed.data;
}

/**
 * Backend origin for Link preconnect hints (no trailing slash).
 */
export function getBackendOrigin(): string {
  const { BOMS_BACKEND_URL } = getServerEnv();
  return BOMS_BACKEND_URL.replace(/\/$/, "");
}
