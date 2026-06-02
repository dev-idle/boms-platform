import { randomUUID } from "node:crypto";

import { NextResponse } from "next/server";
import type { NextProxy, NextRequest } from "next/server";

import { AUTH_REFRESH_COOKIE } from "@/constants/cookies";
import { ROUTE } from "@/constants/routes";
import { getBackendOrigin, getServerEnv } from "@/lib/env";
import { isProtectedPath } from "@/lib/routing/role-routes";
import { validateNext } from "@/lib/validate-next";

function isApiPath(pathname: string): boolean {
  return pathname.startsWith("/api/");
}

function hasRefreshCookie(request: NextRequest): boolean {
  return request.cookies.has(AUTH_REFRESH_COOKIE);
}

function stripUntrustedInboundHeaders(request: NextRequest): Headers {
  const headers = new Headers(request.headers);
  const blocked = [
    "x-internal-secret",
    "x-user-role",
    "x-request-id",
    "x-auth-hint",
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
 * Cookie presence is a coarse gate only — client bootstrap validates the session.
 * Refresh cookie must use Path=/ (see backend AuthCookiePath) so it is sent on page navigations.
 * Role checks happen in client RoleGate / server API RequireRole.
 */
export const proxy: NextProxy = (request) => {
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
  const authHint = hasRefreshCookie(request) ? "present" : "absent";

  // Page navigation: gate protected paths on cookie presence; redirect to login when absent.
  // API calls (/api/*) must always reach the backend so it can answer 401 with the proper envelope —
  // never redirect them.
  if (
    !isApiPath(pathname) &&
    isProtectedPath(pathname) &&
    !hasRefreshCookie(request)
  ) {
    const login = new URL(ROUTE.login, request.url);
    const nextPath = validateNext(`${pathname}${search}`);
    if (nextPath) {
      login.searchParams.set("next", nextPath);
    }
    return NextResponse.redirect(login);
  }

  const requestHeaders = stripUntrustedInboundHeaders(request);
  requestHeaders.set("X-Request-ID", requestId);
  requestHeaders.set("X-Auth-Hint", authHint);
  // Signed shared secret — verified by Fiber middleware on every /api/v1/* request.
  // For non-API navigation it's harmless; keep set everywhere for consistency.
  requestHeaders.set("X-Internal-Secret", env.INTERNAL_PROXY_SECRET);

  const response = NextResponse.next({
    request: { headers: requestHeaders },
  });

  response.headers.set("X-Request-ID", requestId);
  applyResourceHints(response, getBackendOrigin());

  return response;
};

// Include /api/* for request-id / auth-hint consistency. Browser API traffic is handled by
// app/api/v1/[...path]/route.ts (BFF), which stamps X-Internal-Secret server-side.
export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|robots.txt|sitemap.xml|.*\\.(?:svg|png|jpg|jpeg|gif|webp|ico)$).*)",
  ],
};
