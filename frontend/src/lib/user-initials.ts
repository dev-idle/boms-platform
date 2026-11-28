import type { Me } from "@/lib/schemas/me";

/** The one letter every avatar shows: the name's first, else the email's. */
export function initialOf(name: string | null | undefined, email: string): string {
  return (name?.trim() || email).charAt(0).toUpperCase();
}

/** The signed-in account's avatar letter. */
export function userInitial(user: Me | null | undefined): string {
  if (!user) {
    return "?";
  }
  const name = user.profile.type === "customer" ? user.profile.display_name : user.profile.full_name;
  return initialOf(name, user.email);
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
