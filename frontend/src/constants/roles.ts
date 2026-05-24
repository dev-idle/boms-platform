/** Application roles — mirror `backend/internal/domain/user/user.go`. */
export const USER_ROLE = {
  customer: "customer",
  staff: "staff",
  baker: "baker",
  manager: "manager",
  admin: "admin",
} as const;

export type UserRole = (typeof USER_ROLE)[keyof typeof USER_ROLE];

/** Roles allowed to access admin dashboard routes. */
export const ADMIN_ROLES: readonly UserRole[] = [
  USER_ROLE.admin,
  USER_ROLE.manager,
] as const;

/** Roles using the operational staff profile. */
export const STAFF_ROLES: readonly UserRole[] = [
  USER_ROLE.staff,
  USER_ROLE.baker,
  USER_ROLE.manager,
] as const;
