import type { DashboardBreadcrumbItem } from "@/components/ui/dashboard-breadcrumb";
import { ROUTE } from "@/constants/routes";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

/** Admin namespace — Users section trails (sidebar "Users" is the section root). */

export function adminUsersBreadcrumb(
  leaf: string,
): DashboardBreadcrumbItem[] {
  return [
    { label: PAGE_TITLES.users, href: ROUTE.admin.users },
    { label: leaf },
  ];
}

export function adminUserDetailBreadcrumbItems(): DashboardBreadcrumbItem[] {
  return adminUsersBreadcrumb(PAGE_TITLES.breadcrumbDetail);
}
