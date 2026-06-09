import { USER_ROLE, type UserRole } from "@/constants/roles";
import { GUEST_STOREFRONT_ROUTE_PREFIXES, ROUTE } from "@/constants/routes";

function matchesPrefix(pathname: string, prefix: string): boolean {
  return pathname === prefix || pathname.startsWith(`${prefix}/`);
}

export function isGuestStorefrontPath(pathname: string): boolean {
  return GUEST_STOREFRONT_ROUTE_PREFIXES.some((prefix) =>
    matchesPrefix(pathname, prefix),
  );
}

export function isPublicAuthEntryPath(pathname: string): boolean {
  return pathname === ROUTE.login || pathname === ROUTE.register;
}

/** Customer may browse home + catalog while signed in. */
export function allowsAuthenticatedCustomerPublicBrowsing(
  pathname: string,
): boolean {
  return pathname === ROUTE.home || isGuestStorefrontPath(pathname);
}

export function shouldRedirectAuthenticatedPublicUser(
  pathname: string,
  role: UserRole,
): boolean {
  if (isPublicAuthEntryPath(pathname)) {
    return true;
  }
  if (role === USER_ROLE.customer) {
    return false;
  }
  return pathname === ROUTE.home || isGuestStorefrontPath(pathname);
}
