import "server-only";

import { cookies } from "next/headers";
import type { ZodType } from "zod";

import { getServerEnv } from "@/lib/env";
import { BomsApiError, BomsValidationError } from "@/types/api";

type Json = Record<string, unknown> | unknown[] | string | number | boolean | null;

export type FiberRequestInit = Omit<RequestInit, "body"> & {
  /** Zod schema for JSON response body (enforced at trust boundary). */
  schema?: ZodType<unknown>;
  /** JSON body for non-GET requests. */
  json?: Json;
  /** Skip attaching browser cookies (rare; prefer default forwarding). */
  skipCookieForwarding?: boolean;
};

const REQUEST_TIMEOUT_MS = 25_000;

function buildUrl(path: string): string {
  const base = getServerEnv().BOMS_BACKEND_URL.replace(/\/$/, "");
  const p = path.startsWith("/") ? path : `/${path}`;
  return `${base}${p}`;
}

async function parseJsonSafe(response: Response): Promise<unknown> {
  const text = await response.text();
  if (!text) {
    return null;
  }
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return { raw: text };
  }
}

export class BomsApiClient {
  async request<T = unknown>(
    path: string,
    init: FiberRequestInit = {},
  ): Promise<T> {
    const { schema, json, skipCookieForwarding, ...rest } = init;
    const headers = new Headers(rest.headers);

    if (!headers.has("Accept")) {
      headers.set("Accept", "application/json");
    }

    if (!skipCookieForwarding) {
      const jar = await cookies();
      const serialized = jar
        .getAll()
        .map((c) => `${c.name}=${c.value}`)
        .join("; ");
      if (serialized) {
        headers.set("Cookie", serialized);
      }
    }

    if (json !== undefined) {
      headers.set("Content-Type", "application/json");
    }

    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);

    let response: Response;
    try {
      response = await fetch(buildUrl(path), {
        ...rest,
        headers,
        body: json !== undefined ? JSON.stringify(json) : undefined,
        signal: rest.signal ?? controller.signal,
        cache: "no-store",
      });
    } catch (e) {
      if (e instanceof Error && e.name === "AbortError") {
        throw new BomsApiError("Upstream request timed out", 504);
      }
      throw e;
    } finally {
      clearTimeout(timeout);
    }

    const payload = await parseJsonSafe(response);

    if (!response.ok) {
      throw new BomsApiError(
        typeof payload === "object" &&
          payload !== null &&
          "message" in payload &&
          typeof (payload as { message?: unknown }).message === "string"
          ? (payload as { message: string }).message
          : `Fiber responded with HTTP ${response.status}`,
        response.status,
        payload,
      );
    }

    if (schema) {
      const parsed = schema.safeParse(payload);
      if (!parsed.success) {
        throw new BomsValidationError(
          "Response failed Zod validation at DAL boundary",
          parsed.error.flatten(),
        );
      }
      return parsed.data as T;
    }

    return payload as T;
  }
}

let singleton: BomsApiClient | null = null;

/**
 * Singleton HTTP client for Go Fiber. Import only from `src/lib/dal/*`.
 */
export function getBomsApiClient(): BomsApiClient {
  if (!singleton) {
    singleton = new BomsApiClient();
  }
  return singleton;
}
