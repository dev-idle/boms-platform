import { USER_ROLE } from "@/constants/roles";
import { initialOf } from "@/lib/user-initials";

import type { AdminUser } from "../schemas";

export function adminUserDisplayName(user: AdminUser): string {
  if (user.role === USER_ROLE.customer) {
    return user.display_name?.trim() || user.email;
  }

  return user.full_name?.trim() || user.email.split("@")[0] || user.email;
}

function profileName(user: AdminUser): string | undefined {
  const name = user.role === USER_ROLE.customer ? user.display_name : user.full_name;
  return name?.trim() || undefined;
}

/**
 * The line under an account's email in the users table: the most useful fact
 * there is about it, so the line is never blank.
 */
export function adminUserAccountNote(user: AdminUser, currentUserId: string | undefined): string {
  const name = profileName(user);
  if (name) {
    return user.employee_code ? `${name} · ${user.employee_code}` : name;
  }
  if (!user.email_verified) {
    return "Email not verified";
  }
  return user.id === currentUserId ? "Admin — this is you" : "No name on file";
}

/** The admin table's avatar letter. */
export function adminUserInitial(user: AdminUser): string {
  return initialOf(profileName(user), user.email);
}

