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
    <Suspense fallback={null}>
      <ProductCatalog />
      <ComboCatalogSection />
    </Suspense>
  );
}
