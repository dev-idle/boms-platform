import { randomUUID } from "node:crypto";

import { NextResponse } from "next/server";
import type { NextProxy, NextRequest } from "next/server";

import {
  AUTH_ACCESS_COOKIE,
  AUTH_REFRESH_COOKIE,
  ROLE_COOKIE,
} from "@/constants/cookies";
import {
  PROTECTED_ROUTE_PREFIXES,
  PUBLIC_AUTH_ROUTE_PREFIXES,
  ROUTE,
} from "@/constants/routes";
import { getBackendOrigin, getServerEnv } from "@/lib/env";

function isProtectedPath(pathname: string): boolean {
  return PROTECTED_ROUTE_PREFIXES.some(
    (prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`),
  );
}

function isPublicAuthPath(pathname: string): boolean {
  return PUBLIC_AUTH_ROUTE_PREFIXES.some(
    (p) => pathname === p || pathname.startsWith(`${p}/`),
  );
}

function hasAuthCookies(request: NextRequest): boolean {
  return (
    request.cookies.has(AUTH_ACCESS_COOKIE) &&
    request.cookies.has(AUTH_REFRESH_COOKIE)
  );
}

function resolveUserRole(request: NextRequest): string {
  const raw = request.cookies.get(ROLE_COOKIE)?.value ?? "";
  if (/^[a-z0-9:_-]{1,64}$/i.test(raw)) {
    return raw;
  }
  return hasAuthCookies(request) ? "authenticated" : "anonymous";
}

function stripUntrustedInboundHeaders(request: NextRequest): Headers {
  const headers = new Headers(request.headers);
  const blocked = [
    "x-internal-secret",
    "x-user-role",
    "x-request-id",
  ] as const;
  for (const name of blocked) {
    headers.delete(name);
  }
  return headers;
}

function applyResourceHints(response: NextResponse, backendOrigin: string): void {
  const origin = backendOrigin.replaceAll(/[\r\n"]/g, "");
  response.headers.append("Link", `<${origin}>; rel=preconnect`);
}

/**
 * Network boundary for BOMS. Runs before route rendering (Node runtime by default).
 *
 * HTTP 103 Early Hints are not emitted here because `NextResponse` does not expose a
 * portable multi-status composition path across hosts. `Link: rel=preconnect` hints
 * warm browser connection pools toward Fiber similarly to CDN Early Hints for APIs.
 */
export const proxy: NextProxy = (request, event) => {
  const { pathname, search } = request.nextUrl;

  if (pathname.startsWith("/_next") || pathname.startsWith("/favicon.ico")) {
    return NextResponse.next();
  }

  let env: ReturnType<typeof getServerEnv>;
  try {
    env = getServerEnv();
  } catch {
    return NextResponse.json(
      { error: "server_misconfigured", detail: "Missing or invalid BOMS env" },
      { status: 503 },
    );
  }

  const requestId = randomUUID();
  const role = resolveUserRole(request);

  if (isProtectedPath(pathname) && !hasAuthCookies(request)) {
    const login = new URL(ROUTE.login, request.url);
    login.searchParams.set("next", `${pathname}${search}`);
    return NextResponse.redirect(login);
  }

  if (isPublicAuthPath(pathname) && hasAuthCookies(request)) {
    return NextResponse.redirect(new URL(ROUTE.home, request.url));
  }

  const requestHeaders = stripUntrustedInboundHeaders(request);
  requestHeaders.set("X-Request-ID", requestId);
  requestHeaders.set("X-User-Role", role);
  requestHeaders.set("X-Internal-Secret", env.INTERNAL_PROXY_SECRET);

  const response = NextResponse.next({
    request: { headers: requestHeaders },
  });

  response.headers.set("X-Request-ID", requestId);
  applyResourceHints(response, getBackendOrigin());

  const warmUrl = `${getBackendOrigin()}/health`;
  event.waitUntil(
    fetch(warmUrl, {
      method: "HEAD",
      headers: {
        "X-Internal-Secret": env.INTERNAL_PROXY_SECRET,
        "X-Request-ID": requestId,
      },
    }).catch(() => undefined),
  );

  return response;
};

export const config = {
  matcher: [
    "/((?!api|_next/static|_next/image|favicon.ico|robots.txt|sitemap.xml|.*\\.(?:svg|png|jpg|jpeg|gif|webp|ico)$).*)",
  ],
};
