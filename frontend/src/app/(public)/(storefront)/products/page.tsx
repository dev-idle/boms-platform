import { Suspense } from "react";

import { PageLoadingState } from "@/components/ui/loading-state";
import { ProductCatalog } from "@/features/catalog/components/product-catalog";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

import { ComboCatalogSection } from "./combo-catalog-section";

export const metadata = pageTitle(
  PAGE_TITLES.shop,
  "Browse pastries, cakes, and seasonal treats for pickup.",
);

export default function ProductsPage() {
  return (
    <Suspense
      fallback={
        <div className="storefront-container py-10">
          <PageLoadingState />
        </div>
      }
    >
      <ProductCatalog />
      <ComboCatalogSection />
    </Suspense>
  );
}
