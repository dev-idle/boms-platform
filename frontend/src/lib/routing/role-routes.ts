import { USER_ROLE, type UserRole } from "@/constants/roles";
import {
  ADMIN_ROUTE_PREFIXES,
  BAKER_ROUTE_PREFIXES,
  CUSTOMER_ROUTE_PREFIXES,
  MANAGER_ROUTE_PREFIXES,
  PROTECTED_ROUTE_PREFIXES,
  ROUTE,
  STAFF_ROUTE_PREFIXES,
} from "@/constants/routes";

function assertNeverRole(role: never): never {
  throw new Error(`Unhandled role: ${role}`);
}

export function routePrefixesForRole(role: UserRole): readonly string[] {
  switch (role) {
    case USER_ROLE.customer:
      return CUSTOMER_ROUTE_PREFIXES;
    case USER_ROLE.staff:
      return STAFF_ROUTE_PREFIXES;
    case USER_ROLE.baker:
      return BAKER_ROUTE_PREFIXES;
    case USER_ROLE.manager:
      return MANAGER_ROUTE_PREFIXES;
    case USER_ROLE.admin:
      return ADMIN_ROUTE_PREFIXES;
    default:
      return assertNeverRole(role);
  }
}

/** Default landing path after sign-in for each role. */
export function homeRouteForRole(role: UserRole): string {
  switch (role) {
    case USER_ROLE.customer:
      return ROUTE.products;
    case USER_ROLE.staff:
      return ROUTE.staff.account.profile;
    case USER_ROLE.baker:
      return ROUTE.baker.account.profile;
    case USER_ROLE.manager:
      return ROUTE.manager.categories;
    case USER_ROLE.admin:
      return ROUTE.admin.dashboard;
    default:
      return assertNeverRole(role);
  }
}

export function passwordRouteForRole(role: UserRole): string {
  switch (role) {
    case USER_ROLE.customer:
      return ROUTE.customer.account.password;
    case USER_ROLE.staff:
      return ROUTE.staff.account.password;
    case USER_ROLE.baker:
      return ROUTE.baker.account.password;
    case USER_ROLE.manager:
      return ROUTE.manager.account.password;
    case USER_ROLE.admin:
      // Admin change-password UI lives on the profile page (no `/admin/account/password`).
      return ROUTE.admin.account.profile;
    default:
      return assertNeverRole(role);
  }
}

export function profileRouteForRole(role: UserRole): string {
  switch (role) {
    case USER_ROLE.customer:
      return ROUTE.customer.account.profile;
    case USER_ROLE.staff:
      return ROUTE.staff.account.profile;
    case USER_ROLE.baker:
      return ROUTE.baker.account.profile;
    case USER_ROLE.manager:
      return ROUTE.manager.account.profile;
    case USER_ROLE.admin:
      return ROUTE.admin.account.profile;
    default:
      return assertNeverRole(role);
  }
}

function matchesRoutePrefix(pathname: string, prefix: string): boolean {
  return pathname === prefix || pathname.startsWith(`${prefix}/`);
}

/** True when the path requires an authenticated session (page navigation). */
export function isProtectedPath(pathname: string): boolean {
  return PROTECTED_ROUTE_PREFIXES.some((prefix) =>
    matchesRoutePrefix(pathname, prefix),
  );
}

export function isPathAllowedForRole(pathname: string, role: UserRole): boolean {
  return routePrefixesForRole(role).some((prefix) =>
    matchesRoutePrefix(pathname, prefix),
  );
}
