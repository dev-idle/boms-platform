import "server-only";

import { randomUUID } from "node:crypto";

import { cookies } from "next/headers";
import type { ZodType } from "zod";

import { apiEnvelopeSchema } from "@/lib/api-envelope";
import { getServerEnv } from "@/lib/env";
import { BomsApiError, BomsValidationError } from "@/lib/errors";

type Json = Record<string, unknown> | unknown[] | string | number | boolean | null;

export type FiberRequestInit = Omit<RequestInit, "body"> & {
  /** Zod schema for JSON response data (enforced at trust boundary). */
  schema?: ZodType<unknown>;
  /** JSON body for non-GET requests. */
  json?: Json;
  /** Skip attaching browser cookies (rare; prefer default forwarding). */
  skipCookieForwarding?: boolean;
};

const REQUEST_TIMEOUT_MS = 25_000;
const INTERNAL_SECRET_HEADER = "X-Internal-Secret";

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
    return null;
  }
}

function throwFromEnvelope(status: number, payload: unknown): never {
  const envelope = apiEnvelopeSchema.safeParse(payload);
  if (envelope.success && envelope.data.error) {
    throw new BomsApiError(envelope.data.error.message, status, payload);
  }
  throw new BomsApiError(`Request failed with HTTP ${status}`, status, payload);
}

export class BomsApiClient {
  async request<T = unknown>(
    path: string,
    init: FiberRequestInit = {},
  ): Promise<T> {
    const { schema, json, skipCookieForwarding, ...rest } = init;
    const env = getServerEnv();
    const headers = new Headers(rest.headers);

    if (!headers.has("Accept")) {
      headers.set("Accept", "application/json");
    }

    headers.set(INTERNAL_SECRET_HEADER, env.INTERNAL_PROXY_SECRET);

    if (!headers.has("X-Request-ID")) {
      headers.set("X-Request-ID", randomUUID());
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

    if (response.status === 204) {
      return undefined as T;
    }

    const payload = await parseJsonSafe(response);

    if (!response.ok) {
      throwFromEnvelope(response.status, payload);
    }

    const envelope = apiEnvelopeSchema.safeParse(payload);
    if (!envelope.success) {
      throw new BomsApiError("Response failed envelope validation", 502, payload);
    }

    if (!envelope.data.success) {
      throwFromEnvelope(response.status, payload);
    }

    const data = envelope.data.data;

    if (schema) {
      const parsed = schema.safeParse(data);
      if (!parsed.success) {
        throw new BomsValidationError(
          "Response failed Zod validation at DAL boundary",
          parsed.error.flatten(),
        );
      }
      return parsed.data as T;
    }

    return data as T;
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
