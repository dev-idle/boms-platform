import { USER_ROLE } from "@/constants/roles";

import type { AdminUser } from "../schemas";

export function adminUserDisplayName(user: AdminUser): string {
  if (user.role === USER_ROLE.customer) {
    return user.display_name?.trim() || user.email;
  }

  return user.full_name?.trim() || user.email.split("@")[0] || user.email;
}

/** Table name column — empty profile fields render as em dash. */
export function adminUserListName(user: AdminUser): string {
  if (user.role === USER_ROLE.customer) {
    return user.display_name?.trim() || "—";
  }

  return user.full_name?.trim() || "—";
}

