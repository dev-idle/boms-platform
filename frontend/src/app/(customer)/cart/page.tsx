import { StorefrontPageHeader } from "@/components/layouts/storefront-page-header";
import { CartView } from "@/features/customer";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.cart);

export default function CartPage() {
  return (
    <div className="storefront-page-stack">
      <StorefrontPageHeader title={PAGE_TITLES.cart} />
      <CartView />
    </div>
  );
}
