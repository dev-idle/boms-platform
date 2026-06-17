import { OrderList } from "@/features/customer";
import { StorefrontPageHeader } from "@/components/layouts/storefront-page-header";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.orders);

export default function OrdersPage() {
  return (
    <div className="storefront-customer-section">
      <StorefrontPageHeader
        lead="Your order history and pickup totals."
        title={PAGE_TITLES.orders}
      />
      <OrderList />
    </div>
  );
}
