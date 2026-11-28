import { StorefrontAccountShell } from "@/components/layouts/storefront-account-shell";
import { SavedProducts } from "@/features/customer";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.saved);

export default function SavedPage() {
  return (
    <StorefrontAccountShell title={PAGE_TITLES.saved}>
      <SavedProducts />
    </StorefrontAccountShell>
  );
}
