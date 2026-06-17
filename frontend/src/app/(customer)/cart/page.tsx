import { CartView } from "@/features/customer";
import { StorefrontPageHeader } from "@/components/layouts/storefront-page-header";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.cart);

export default function CartPage() {
  return (
    <div className="storefront-customer-section">
      <StorefrontPageHeader
        lead="Prices and discounts are calculated on the server when your cart loads or at checkout."
        title={PAGE_TITLES.cart}
      />
      <CartView />
    </div>
  );
}
