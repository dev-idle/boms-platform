import { randomUUID } from "node:crypto";

import { NextResponse } from "next/server";
import type { NextProxy, NextRequest } from "next/server";

import { AUTH_REFRESH_COOKIE } from "@/constants/cookies";
import { PROTECTED_ROUTE_PREFIXES, ROUTE } from "@/constants/routes";
import { getBackendOrigin, getServerEnv } from "@/lib/env";
import { validateNext } from "@/lib/validate-next";

function isProtectedPath(pathname: string): boolean {
  return PROTECTED_ROUTE_PREFIXES.some(
    (prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`),
  );
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

  if (isProtectedPath(pathname) && !hasRefreshCookie(request)) {
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
  requestHeaders.set("X-Internal-Secret", env.INTERNAL_PROXY_SECRET);

  const response = NextResponse.next({
    request: { headers: requestHeaders },
  });

  response.headers.set("X-Request-ID", requestId);
  applyResourceHints(response, getBackendOrigin());

  return response;
};

export const config = {
  matcher: [
    "/((?!api|_next/static|_next/image|favicon.ico|robots.txt|sitemap.xml|.*\\.(?:svg|png|jpg|jpeg|gif|webp|ico)$).*)",
  ],
};
