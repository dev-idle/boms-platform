import { StaffOrdersTable } from "@/features/staff";

export default function StaffOrdersPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">
          Orders
        </h1>
        <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
          Review customer orders and update their status.
        </p>
      </div>
      <StaffOrdersTable />
    </div>
  );
}
