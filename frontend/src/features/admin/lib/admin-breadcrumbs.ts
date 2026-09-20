import type { DashboardBreadcrumbItem } from "@/components/ui/dashboard-breadcrumb";
import { ROUTE } from "@/constants/routes";

/** Admin namespace — Users trails. The root crumb matches the sidebar item ("Users"). */
const USERS_CRUMB_LABEL = "Users";

export function adminUsersBreadcrumb(
  leaf: string,
): DashboardBreadcrumbItem[] {
  return [
    { label: USERS_CRUMB_LABEL, href: ROUTE.admin.users },
    { label: leaf },
  ];
}

/** Detail trail ends in the person, so the admin sees where they are. */
export function adminUserDetailBreadcrumbItems(
  displayName: string,
): DashboardBreadcrumbItem[] {
  return adminUsersBreadcrumb(displayName);
}
