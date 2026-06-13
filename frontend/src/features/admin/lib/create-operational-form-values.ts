import { USER_ROLE } from "@/constants/roles";

import type { CreateOperationalInput } from "../schemas";

export const CREATE_OPERATIONAL_INITIAL: CreateOperationalInput = {
  email: "",
  role: USER_ROLE.staff,
  full_name: "",
  phone: null,
  employee_code: "",
};
