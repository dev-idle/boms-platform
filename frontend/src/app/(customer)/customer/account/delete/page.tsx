import { USER_ROLE } from "@/constants/roles";
import { deleteAccountRouteForRole } from "@/lib/routing/role-routes";
import { redirect } from "next/navigation";

export default function CustomerAccountDeletePage() {
  redirect(deleteAccountRouteForRole(USER_ROLE.customer));
}
