import { USER_ROLE } from "@/constants/roles";
import { passwordRouteForRole } from "@/lib/routing/role-routes";
import { redirect } from "next/navigation";

export default function CustomerAccountPasswordPage() {
  redirect(passwordRouteForRole(USER_ROLE.customer));
}
