/** Application roles — mirror `backend/internal/domain/user/role.go`. */
export const USER_ROLE = {
  customer: "customer",
  staff: "staff",
  baker: "baker",
  manager: "manager",
  admin: "admin",
} as const;

export type UserRole = (typeof USER_ROLE)[keyof typeof USER_ROLE];

/** Operational roles an admin may assign (excludes customer and admin). */
export const ASSIGNABLE_OPERATIONAL_ROLES: readonly UserRole[] = [
  USER_ROLE.staff,
  USER_ROLE.baker,
  USER_ROLE.manager,
] as const;

/** Title-case label for UI (API values stay lowercase). */
export function roleDisplayLabel(role: UserRole): string {
  return role.charAt(0).toUpperCase() + role.slice(1);
}
