import type { UserRole } from "@/constants/roles";
import { ROUTE } from "@/constants/routes";
import { isPathAllowedForRole } from "@/lib/routing/role-routes";

/**
 * Accepts relative in-app paths only. Rejects open-redirect vectors.
 */
export function validateNext(next: string | null | undefined): string | null {
  if (!next) {
    return null;
  }

  let decoded = next.trim();
  try {
    decoded = decodeURIComponent(decoded);
  } catch {
    return null;
  }

  if (!decoded.startsWith("/")) {
    return null;
  }
  if (decoded.startsWith("//")) {
    return null;
  }
  if (/[\u0000-\u001F\u007F]/.test(decoded)) {
    return null;
  }
  if (decoded.includes("@")) {
    return null;
  }

  const lower = decoded.toLowerCase();
  if (
    lower.includes("http://") ||
    lower.includes("https://") ||
    lower.includes("%2f%2f") ||
    lower.includes("%5c") ||
    decoded.includes("\\")
  ) {
    return null;
  }

  return decoded;
}

/** Same as validateNext, but the path must belong to the signed-in user's role namespace. */
export function validateNextForRole(
  next: string | null | undefined,
  role: UserRole,
): string | null {
  const path = validateNext(next);
  if (!path) {
    return null;
  }
  if (!isPathAllowedForRole(path, role)) {
    return null;
  }
  return path;
}

function authHrefPreservingNext(
  route: string,
  pathname: string,
  search = "",
): string {
  const next = validateNext(`${pathname}${search}`);
  if (!next) {
    return route;
  }
  return `${route}?next=${encodeURIComponent(next)}`;
}

/** Safe login URL that preserves the current protected path for post-auth redirect. */
export function loginHrefPreservingNext(
  pathname: string,
  search = "",
): string {
  return authHrefPreservingNext(ROUTE.login, pathname, search);
}

/** Safe register URL that preserves the current path for post-auth redirect. */
export function registerHrefPreservingNext(
  pathname: string,
  search = "",
): string {
  return authHrefPreservingNext(ROUTE.register, pathname, search);
}

/** Login URL with a validated `next` query (e.g. after registration). */
export function loginHrefWithNext(next: string | null | undefined): string {
  const safeNext = validateNext(next);
  if (!safeNext) {
    return ROUTE.login;
  }
  return `${ROUTE.login}?next=${encodeURIComponent(safeNext)}`;
}

/** Register URL with a validated `next` query (e.g. from login hop). */
export function registerHrefWithNext(next: string | null | undefined): string {
  const safeNext = validateNext(next);
  if (!safeNext) {
    return ROUTE.register;
  }
  return `${ROUTE.register}?next=${encodeURIComponent(safeNext)}`;
}

/** Login URL after registration, preserving optional post-auth redirect. */
export function loginHrefAfterRegister(
  next: string | null | undefined,
): string {
  const params = new URLSearchParams({ registered: "1" });
  const safeNext = validateNext(next);
  if (safeNext) {
    params.set("next", safeNext);
  }
  return `${ROUTE.login}?${params.toString()}`;
}
