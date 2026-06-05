import { OrderList } from "@/features/customer";

export default function OrdersPage() {
  return (
    <div>
      <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">Orders</h1>
      <p className="mt-2 max-w-prose text-sm text-zinc-600 dark:text-zinc-400">
        Your order history and totals.
      </p>
      <div className="mt-6">
        <OrderList />
      </div>
    </div>
  );
}
