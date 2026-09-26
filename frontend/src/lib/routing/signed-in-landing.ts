import { USER_ROLE, type UserRole } from "@/constants/roles";

import {
  allowsAuthenticatedCustomerPublicBrowsing,
  isPublicAuthEntryPath,
  shouldRedirectAuthenticatedPublicUser,
} from "./guest-storefront";
import { homeRouteForRole, isPathAllowedForRole, isProtectedPath } from "./role-routes";
import { validateNextForRole } from "../validate-next";

/** The role a session cookie claims, or undefined when it claims nothing usable. */
export function parseRoleHint(value: string | undefined): UserRole | undefined {
  const roles = Object.values(USER_ROLE) as string[];
  return value && roles.includes(value) ? (value as UserRole) : undefined;
}

/**
 * Where a signed-in visitor should land, decided before any HTML is sent.
 *
 * The access token lives in memory, so without this the browser has to restore
 * the session before it can know the role — and a manager opening the site would
 * watch the storefront render, then jump to the dashboard a round trip later.
 *
 * The role comes from a cookie, which is a hint and never authority: it decides
 * only *where* the visitor is sent. Every page behind it is still gated by the
 * session itself, and every byte of data by the API.
 *
 * Returns null when the visitor is already somewhere they may be.
 */
export function signedInLanding(
  pathname: string,
  search: string,
  role: UserRole,
): string | null {
  if (isPublicAuthEntryPath(pathname)) {
    // A deep link the visitor followed before signing in still wins, as long as
    // it belongs to them.
    const next = new URLSearchParams(search).get("next");
    return validateNextForRole(next, role) ?? homeRouteForRole(role);
  }

  if (isProtectedPath(pathname)) {
    return isPathAllowedForRole(pathname, role) ? null : homeRouteForRole(role);
  }

  if (allowsAuthenticatedCustomerPublicBrowsing(pathname)) {
    return shouldRedirectAuthenticatedPublicUser(pathname, role)
      ? homeRouteForRole(role)
      : null;
  }

  return null;
}
