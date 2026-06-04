import { USER_ROLE } from "@/constants/roles";

import type {
  AdminUser,
  AdminUserProfileFormValues,
  UpdateRoleInput,
} from "../schemas";

export function adminUserToProfileFormValues(
  user: AdminUser,
): AdminUserProfileFormValues {
  return {
    full_name: user.full_name ?? "",
    phone: user.phone ?? "",
    employee_code: user.employee_code ?? "",
    hire_date: user.hire_date ? user.hire_date.slice(0, 10) : "",
    shift: user.shift ?? "",
  };
}

export function adminUserToRoleFormValues(user: AdminUser): UpdateRoleInput {
  const roleForForm =
    user.role === USER_ROLE.staff ||
    user.role === USER_ROLE.baker ||
    user.role === USER_ROLE.manager
      ? user.role
      : USER_ROLE.staff;

  return {
    role: roleForForm,
    full_name: user.full_name ?? "",
    phone: user.phone ?? "",
    employee_code: user.employee_code ?? "",
    hire_date: user.hire_date ? user.hire_date.slice(0, 10) : "",
    shift: user.shift ?? "",
  };
}
