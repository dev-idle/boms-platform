import { Suspense } from "react";

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
        <div className="mx-auto max-w-7xl px-4 py-10 text-sm text-muted sm:px-6 lg:px-8">
          Loading shop…
        </div>
      }
    >
      <ProductCatalog />
      <ComboCatalogSection />
    </Suspense>
  );
}
