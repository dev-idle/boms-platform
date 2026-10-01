import Link from "next/link";

import { Button } from "@/components/ui/button";
import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { DASHBOARD_PAGE_EYEBROW } from "@/constants/dashboard-page-copy";
import { ROUTE } from "@/constants/routes";
import { StaffOrdersTable } from "@/features/staff";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.orders);

export default function StaffOrdersPage() {
  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        actions={
          <Button asChild>
            <Link href={ROUTE.staff.newOrder}>+ New order</Link>
          </Button>
        }
        description="Review customer orders and update their status."
        eyebrow={DASHBOARD_PAGE_EYEBROW.operations}
        leadAside
        title={PAGE_TITLES.orders}
      />
      <StaffOrdersTable />
    </div>
  );
}
