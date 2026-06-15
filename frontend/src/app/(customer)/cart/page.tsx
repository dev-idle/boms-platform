import { CartView } from "@/features/customer";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.cart);

export default function CartPage() {
  return (
    <div>
      <h1 className="text-page-title">{PAGE_TITLES.cart}</h1>
      <p className="mt-2 max-w-prose text-sm text-muted">
        Prices and discounts are calculated on the server when your cart loads or
        at checkout.
      </p>
      <div className="mt-6">
        <CartView />
      </div>
    </div>
  );
}
