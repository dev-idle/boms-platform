import { redirect } from "next/navigation";

import { USER_ROLE } from "@/constants/roles";
import { passwordRouteForRole } from "@/lib/routing/role-routes";

export default function AdminAccountPasswordPage() {
  redirect(passwordRouteForRole(USER_ROLE.admin));
}
