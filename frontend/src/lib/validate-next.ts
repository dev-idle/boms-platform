import type { UserRole } from "@/constants/roles";
import { isPathAllowedForRole } from "@/features/user/lib/role-routes";

/**
 * Accepts relative in-app paths only. Rejects open-redirect vectors.
 */
export function validateNext(next: string | null | undefined): string | null {
  if (!next) {
    return null;
  }

  let decoded = next.trim();
  try {
    decoded = decodeURIComponent(decoded);
  } catch {
    return null;
  }

  if (!decoded.startsWith("/")) {
    return null;
  }
  if (decoded.startsWith("//")) {
    return null;
  }
  if (/[\u0000-\u001F\u007F]/.test(decoded)) {
    return null;
  }
  if (decoded.includes("@")) {
    return null;
  }

  const lower = decoded.toLowerCase();
  if (
    lower.includes("http://") ||
    lower.includes("https://") ||
    lower.includes("%2f%2f") ||
    lower.includes("%5c") ||
    decoded.includes("\\")
  ) {
    return null;
  }

  return decoded;
}

/** Same as validateNext, but the path must belong to the signed-in user's role namespace. */
export function validateNextForRole(
  next: string | null | undefined,
  role: UserRole,
): string | null {
  const path = validateNext(next);
  if (!path) {
    return null;
  }
  if (!isPathAllowedForRole(path, role)) {
    return null;
  }
  return path;
}
