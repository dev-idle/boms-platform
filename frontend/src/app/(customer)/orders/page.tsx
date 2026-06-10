import { OrderList } from "@/features/customer";

export default function OrdersPage() {
  return (
    <div>
      <h1 className="text-2xl font-medium text-ink">Orders</h1>
      <p className="mt-2 max-w-prose text-sm text-ink-2">
        Your order history and totals.
      </p>
      <div className="mt-6">
        <OrderList />
      </div>
    </div>
  );
}
