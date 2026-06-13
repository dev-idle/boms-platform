import type { DashboardBreadcrumbItem } from "@/components/ui/dashboard-breadcrumb";
import { ROUTE } from "@/constants/routes";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

/** Admin namespace — Users section trails (sidebar "Users" is the section root). */

export function adminUserDetailBreadcrumbItems(): DashboardBreadcrumbItem[] {
  return [
    { label: PAGE_TITLES.users, href: ROUTE.admin.users },
    { label: PAGE_TITLES.breadcrumbDetail },
  ];
}

export function adminUsersNewBreadcrumbItems(): DashboardBreadcrumbItem[] {
  return [
    { label: PAGE_TITLES.users, href: ROUTE.admin.users },
    { label: PAGE_TITLES.breadcrumbNew },
  ];
}
