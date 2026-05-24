import { USER_ROLE, type UserRole } from "@/constants/roles";
import { ROUTE } from "@/constants/routes";

function assertNeverRole(role: never): never {
  throw new Error(`Unhandled role: ${role}`);
}

export function passwordRouteForRole(role: UserRole): string {
  switch (role) {
    case USER_ROLE.customer:
      return ROUTE.customer.account.password;
    case USER_ROLE.staff:
    case USER_ROLE.baker:
    case USER_ROLE.manager:
      return ROUTE.staff.account.password;
    case USER_ROLE.admin:
      return ROUTE.admin.account.password;
    default:
      return assertNeverRole(role);
  }
}

export function profileRouteForRole(role: UserRole): string {
  switch (role) {
    case USER_ROLE.customer:
      return ROUTE.customer.account.profile;
    case USER_ROLE.staff:
    case USER_ROLE.baker:
    case USER_ROLE.manager:
      return ROUTE.staff.account.profile;
    case USER_ROLE.admin:
      return ROUTE.admin.account.profile;
    default:
      return assertNeverRole(role);
  }
}
