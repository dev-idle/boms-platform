import "server-only";

import { randomUUID } from "node:crypto";

import { cookies } from "next/headers";
import { connection } from "next/server";

import { getServerEnv } from "@/lib/env";

const INTERNAL_SECRET_HEADER = "X-Internal-Secret";
const REQUEST_TIMEOUT_MS = 25_000;

/** Client must never supply these; the BFF re-stamps trusted values. */
const STRIP_REQUEST_HEADERS = new Set([
  "connection",
  "content-length",
  "host",
  "keep-alive",
  "te",
  "trailer",
  "transfer-encoding",
  "upgrade",
  "x-auth-hint",
  "x-internal-secret",
  "x-request-id",
  "x-user-role",
]);

function buildBackendUrl(pathSegments: string[], search: string): string {
  const env = getServerEnv();
  const base = env.BOMS_BACKEND_URL.replace(/\/$/, "");
  const path = pathSegments.map(encodeURIComponent).join("/");
  const suffix = search.length > 0 ? search : "";
  return `${base}/api/v1/${path}${suffix}`;
}

async function serializeRequestCookies(): Promise<string | undefined> {
  const jar = await cookies();
  const serialized = jar
    .getAll()
    .map((c) => `${c.name}=${c.value}`)
    .join("; ");
  return serialized.length > 0 ? serialized : undefined;
}

function forwardResponseHeaders(
  target: Headers,
  backend: Headers,
): void {
  backend.forEach((value, key) => {
    const lower = key.toLowerCase();
    if (lower === "set-cookie" || lower === "transfer-encoding") {
      return;
    }
    target.set(key, value);
  });

  const setCookies =
    typeof backend.getSetCookie === "function"
      ? backend.getSetCookie()
      : [];
  for (const cookie of setCookies) {
    target.append("Set-Cookie", cookie);
  }
}

/**
 * Server-side BFF proxy for browser `/api/v1/*` calls.
 * Injects X-Internal-Secret and forwards HttpOnly cookies — rewrites alone do not
 * guarantee middleware headers reach the upstream Fiber API.
 */
export async function proxyRequestToBackend(
  request: Request,
  pathSegments: string[],
): Promise<Response> {
  await connection();
  const env = getServerEnv();
  const incoming = new URL(request.url);
  const backendUrl = buildBackendUrl(pathSegments, incoming.search);

  const headers = new Headers();
  request.headers.forEach((value, key) => {
    if (!STRIP_REQUEST_HEADERS.has(key.toLowerCase())) {
      headers.set(key, value);
    }
  });

  headers.set(INTERNAL_SECRET_HEADER, env.INTERNAL_PROXY_SECRET);
  if (!headers.has("X-Request-ID")) {
    headers.set("X-Request-ID", randomUUID());
  }

  const cookieHeader = await serializeRequestCookies();
  if (cookieHeader) {
    headers.set("Cookie", cookieHeader);
  }

  const method = request.method.toUpperCase();
  const hasBody = method !== "GET" && method !== "HEAD";

  const init: RequestInit & { duplex?: "half" } = {
    method,
    headers,
    cache: "no-store",
    signal: AbortSignal.timeout(REQUEST_TIMEOUT_MS),
  };

  if (hasBody) {
    init.body = request.body;
    init.duplex = "half";
  }

  let backendResponse: Response;
  try {
    backendResponse = await fetch(backendUrl, init);
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Upstream request failed";
    return Response.json(
      {
        success: false,
        error: { code: "upstream_unreachable", message },
      },
      { status: 502 },
    );
  }

  const responseHeaders = new Headers();
  forwardResponseHeaders(responseHeaders, backendResponse.headers);

  return new Response(backendResponse.body, {
    status: backendResponse.status,
    statusText: backendResponse.statusText,
    headers: responseHeaders,
  });
}
