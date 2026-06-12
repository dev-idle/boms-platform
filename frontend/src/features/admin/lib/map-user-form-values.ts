import { USER_ROLE } from "@/constants/roles";
import { shallowFormValuesEqual } from "@/lib/form-values-equal";

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
  };
}

export function updateRoleFormValuesEqual(
  current: UpdateRoleInput,
  baseline: UpdateRoleInput,
): boolean {
  return shallowFormValuesEqual(current, baseline, {
    nullableKeys: ["phone", "employee_code"],
  });
}
