import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { StaffOrdersTable } from "@/features/staff";

export default function StaffOrdersPage() {
  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        description="Review customer orders and update their status."
        title="Orders"
      />
      <StaffOrdersTable />
    </div>
  );
}
