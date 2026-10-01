import { StorefrontPageHeader } from "@/components/layouts/storefront-page-header";
import { SavedProducts } from "@/features/customer";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.saved);

export default function SavedPage() {
  return (
    <div className="storefront-customer-section">
      <StorefrontPageHeader lead="Your favorites and your wishlist." title={PAGE_TITLES.saved} />
      <SavedProducts />
    </div>
  );
}
