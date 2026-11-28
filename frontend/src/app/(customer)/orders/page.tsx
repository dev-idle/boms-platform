import { StorefrontAccountShell } from "@/components/layouts/storefront-account-shell";
import { OrderList } from "@/features/customer";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.orders);

export default function OrdersPage() {
  return (
    <StorefrontAccountShell title={PAGE_TITLES.orders}>
      <OrderList />
    </StorefrontAccountShell>
  );
}
