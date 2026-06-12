import { USER_ROLE } from "@/constants/roles";
import { shallowFormValuesEqual } from "@/lib/form-values-equal";

import type { CreateOperationalInput } from "../schemas";

export const CREATE_OPERATIONAL_INITIAL: CreateOperationalInput = {
  email: "",
  role: USER_ROLE.staff,
  full_name: "",
  phone: null,
  employee_code: "",
};

export function createOperationalFormValuesEqual(
  current: CreateOperationalInput,
  baseline: CreateOperationalInput,
): boolean {
  return shallowFormValuesEqual(current, baseline, { nullableKeys: ["phone"] });
}
