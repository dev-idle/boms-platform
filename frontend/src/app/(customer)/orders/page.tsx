import { OrderList } from "@/features/customer";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.orders);

export default function OrdersPage() {
  return (
    <div>
      <h1 className="text-page-title">{PAGE_TITLES.orders}</h1>
      <p className="mt-2 max-w-prose text-sm text-ink-2">
        Your order history and totals.
      </p>
      <div className="mt-6">
        <OrderList />
      </div>
    </div>
  );
}
