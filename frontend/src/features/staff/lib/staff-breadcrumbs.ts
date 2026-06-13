import type { DashboardBreadcrumbItem } from "@/components/ui/dashboard-breadcrumb";
import { ROUTE } from "@/constants/routes";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

export function staffOrderDetailBreadcrumbItems(): DashboardBreadcrumbItem[] {
  return [
    { label: PAGE_TITLES.orders, href: ROUTE.staff.orders },
    { label: PAGE_TITLES.breadcrumbDetail },
  ];
}
