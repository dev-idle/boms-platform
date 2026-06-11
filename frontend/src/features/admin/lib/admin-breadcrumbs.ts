import type { DashboardBreadcrumbItem } from "@/components/ui/dashboard-breadcrumb";
import { ROUTE } from "@/constants/routes";

/** Admin namespace — Users section trails (sidebar "Users" is the section root). */

export function adminUserDetailBreadcrumbItems(): DashboardBreadcrumbItem[] {
  return [
    { label: "Users", href: ROUTE.admin.users },
    { label: "Detail" },
  ];
}

export function adminUsersNewBreadcrumbItems(): DashboardBreadcrumbItem[] {
  return [
    { label: "Users", href: ROUTE.admin.users },
    { label: "New user" },
  ];
}
