import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { StaffOrdersTable } from "@/features/staff";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.orders);

export default function StaffOrdersPage() {
  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        description="Review customer orders and update their status."
        title={PAGE_TITLES.orders}
      />
      <StaffOrdersTable />
    </div>
  );
}
