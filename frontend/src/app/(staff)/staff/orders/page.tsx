import { StaffOrdersTable } from "@/features/staff";

export default function StaffOrdersPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-medium text-ink">
          Orders
        </h1>
        <p className="mt-2 text-sm text-ink-2">
          Review customer orders and update their status.
        </p>
      </div>
      <StaffOrdersTable />
    </div>
  );
}
