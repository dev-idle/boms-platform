"use client";

import type { ZodType } from "zod";

import { ApiError, ApiErrorCode, apiErrorFromPayload } from "@/lib/errors";
import { parseApiEnvelope, parseResponseBody } from "@/lib/api-envelope";
import { useAuthStore } from "@/stores/auth-store";

const REQUEST_TIMEOUT_MS = 25_000;

export type BrowserRequestInit = Omit<RequestInit, "body"> & {
  json?: unknown;
  schema?: ZodType<unknown>;
  /** Skip attaching Authorization header (login, register). */
  skipAuth?: boolean;
  /** Skip 401 refresh-and-retry (auth endpoints). */
  skipRefreshRetry?: boolean;
  /** Return successful envelope meta alongside data. */
  withMeta?: boolean;
};

export type BrowserRequestWithMetaResult<T> = {
  data: T;
  meta?: Record<string, unknown>;
};

function isAuthLoopPath(path: string): boolean {
  return (
    path.includes("/auth/refresh") ||
    path.includes("/auth/login") ||
    path.includes("/auth/register")
  );
}

function buildUrl(path: string): string {
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return normalized;
}

async function executeRequest<T>(
  path: string,
  init: BrowserRequestInit,
  isRetry: boolean,
): Promise<T> {
  const { json, schema, skipAuth, skipRefreshRetry, withMeta, ...rest } = init;
  const headers = new Headers(rest.headers);

  if (!headers.has("Accept")) {
    headers.set("Accept", "application/json");
  }

  const sentToken = skipAuth ? null : useAuthStore.getState().accessToken;
  if (sentToken) {
    headers.set("Authorization", `Bearer ${sentToken}`);
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
      credentials: "include",
      signal: rest.signal ?? controller.signal,
      cache: "no-store",
    });
  } catch (error) {
    if (error instanceof Error && error.name === "AbortError") {
      throw new ApiError(504, {
        code: ApiErrorCode.Timeout,
        message: "Request timed out",
      });
    }
    throw error;
  } finally {
    clearTimeout(timeout);
  }

  if (response.status === 401 && !skipRefreshRetry && !isAuthLoopPath(path)) {
    const error = apiErrorFromPayload(response.status, await parseResponseBody(response));
    // Only a rejected access token is worth a refresh. Any other 401 is the
    // answer itself: retrying a wrong current password would send it twice and
    // rotate the refresh token for nothing.
    if (!isRetry && error.isAccessTokenRejected()) {
      const { refreshNow } = await import("@/lib/auth");
      // Refresh only if no other tab has replaced the token this request carried.
      await refreshNow({ redirectOnFailure: false, staleToken: sentToken });
      return executeRequest<T>(path, init, true);
    }
    throw error;
  }

  if (response.status === 204) {
    return undefined as T;
  }

  const payload = await parseResponseBody(response);

  if (!response.ok) {
    throw apiErrorFromPayload(response.status, payload);
  }

  const envelope = parseApiEnvelope(payload);
  if (!envelope.success) {
    throw new ApiError(502, {
      code: ApiErrorCode.InvalidResponse,
      message: "Response failed envelope validation",
    });
  }

  if (!envelope.data.success) {
    throw new ApiError(response.status, envelope.data.error ?? {
      code: ApiErrorCode.Unknown,
      message: "Request failed",
    });
  }

  const data = envelope.data.data;

  if (schema) {
    const parsed = schema.safeParse(data);
    if (!parsed.success) {
      throw new ApiError(502, {
        code: ApiErrorCode.InvalidResponse,
        message: "Response failed schema validation",
      });
    }
    if (withMeta) {
      return {
        data: parsed.data,
        meta: envelope.data.meta as Record<string, unknown> | undefined,
      } as T;
    }
    return parsed.data as T;
  }

  if (withMeta) {
    return {
      data,
      meta: envelope.data.meta as Record<string, unknown> | undefined,
    } as T;
  }

  return data as T;
}

export async function browserRequest<T>(
  path: string,
  init: BrowserRequestInit = {},
): Promise<T> {
  return executeRequest<T>(path, init, false);
}

export async function browserRequestVoid(
  path: string,
  init: BrowserRequestInit = {},
): Promise<void> {
  await executeRequest<void>(path, init, false);
}

export async function browserRequestWithMeta<T>(
  path: string,
  init: BrowserRequestInit = {},
): Promise<BrowserRequestWithMetaResult<T>> {
  return executeRequest<BrowserRequestWithMetaResult<T>>(
    path,
    {
      ...init,
      withMeta: true,
    },
    false,
  );
}
