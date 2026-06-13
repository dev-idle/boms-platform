import type { DashboardBreadcrumbItem } from "@/components/ui/dashboard-breadcrumb";
import { ROUTE } from "@/constants/routes";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

export function managerCategoriesBreadcrumb(
  leaf: string,
): DashboardBreadcrumbItem[] {
  return [
    { label: PAGE_TITLES.categories, href: ROUTE.manager.categories },
    { label: leaf },
  ];
}

export function managerProductsBreadcrumb(leaf: string): DashboardBreadcrumbItem[] {
  return [
    { label: PAGE_TITLES.products, href: ROUTE.manager.products },
    { label: leaf },
  ];
}

export function managerCombosBreadcrumb(leaf: string): DashboardBreadcrumbItem[] {
  return [
    { label: PAGE_TITLES.combos, href: ROUTE.manager.combos },
    { label: leaf },
  ];
}

export function managerDiscountCodesBreadcrumb(
  leaf: string,
): DashboardBreadcrumbItem[] {
  return [
    { label: PAGE_TITLES.discountCodes, href: ROUTE.manager.discountCodes },
    { label: leaf },
  ];
}
