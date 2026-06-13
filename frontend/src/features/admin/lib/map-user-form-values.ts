import { USER_ROLE } from "@/constants/roles";

import type { AdminUser, UpdateRoleInput } from "../schemas";

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
  };
}
