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

/** Up to two letters for the admin table avatar cell; email is the fallback. */
export function adminUserInitials(user: AdminUser): string {
  const name = adminUserListName(user);
  if (name !== "—") {
    const parts = name.split(/\s+/).filter(Boolean);
    if (parts.length >= 2) {
      return `${parts[0]!.slice(0, 1)}${parts[1]!.slice(0, 1)}`.toUpperCase();
    }
    if (parts.length === 1) {
      return parts[0]!.slice(0, 2).toUpperCase();
    }
  }

  return user.email.slice(0, 2).toUpperCase();
}

