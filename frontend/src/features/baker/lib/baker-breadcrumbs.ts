import type { DashboardBreadcrumbItem } from "@/components/ui/dashboard-breadcrumb";
import { ROUTE } from "@/constants/routes";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

export function bakerProductionDetailBreadcrumbItems(orderCode: string): DashboardBreadcrumbItem[] {
  return [{ label: PAGE_TITLES.production, href: ROUTE.baker.production }, { label: orderCode }];
}
