import { Suspense } from "react";

import { ProductCatalogLoading } from "@/features/catalog";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

import { ComboCatalogSection } from "./combo-catalog-section";
import { ProductCatalogSection } from "./product-catalog-section";

export const instant = true;

export const metadata = pageTitle(
  PAGE_TITLES.shop,
  "Browse pastries, cakes, and seasonal treats for pickup.",
);

export default function ProductsPage() {
  return (
    <Suspense fallback={<ProductCatalogLoading />}>
      <ProductCatalogSection />
      <ComboCatalogSection />
    </Suspense>
  );
}
