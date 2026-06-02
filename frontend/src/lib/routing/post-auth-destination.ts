import type { UserRole } from "@/constants/roles";
import { validateNextForRole } from "@/lib/validate-next";

import { homeRouteForRole, passwordRouteForRole } from "./role-routes";

export type PostAuthDestinationOptions = {
  /** Safe relative path from `validateNext` (e.g. login `?next=`). */
  next?: string | null;
  mustChangePassword?: boolean;
};

/** Canonical post-login / returning-session destination for a role. */
export function resolvePostAuthDestination(
  role: UserRole,
  options: PostAuthDestinationOptions = {},
): string {
  if (options.mustChangePassword) {
    return passwordRouteForRole(role);
  }
  return validateNextForRole(options.next, role) ?? homeRouteForRole(role);
}
