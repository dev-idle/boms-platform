import type { Me } from "@/lib/schemas/me";

/** Two-letter initials for dashboard avatar (email fallback). */
export function userInitials(user: Me | null | undefined): string {
  if (!user) {
    return "?";
  }

  if (user.profile.type === "customer") {
    const name = user.profile.display_name?.trim();
    if (name) {
      return name.slice(0, 1).toUpperCase();
    }
  } else {
    const parts = user.profile.full_name.trim().split(/\s+/).filter(Boolean);
    if (parts.length >= 2) {
      return `${parts[0]!.slice(0, 1)}${parts[1]!.slice(0, 1)}`.toUpperCase();
    }
    if (parts.length === 1) {
      return parts[0]!.slice(0, 2).toUpperCase();
    }
  }

  return user.email.slice(0, 1).toUpperCase();
}

/** Display name for dashboard account chrome (sidebar footer). */
export function userDisplayName(user: Me | null | undefined): string {
  if (!user) {
    return "Account";
  }

  if (user.profile.type === "customer") {
    return user.profile.display_name?.trim() || user.email;
  }

  return user.profile.full_name.trim() || user.email;
}
