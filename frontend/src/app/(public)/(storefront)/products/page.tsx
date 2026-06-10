import type { Metadata } from "next";
import { Suspense } from "react";

import { ProductCatalog } from "@/features/catalog/components/product-catalog";

import { ComboCatalogSection } from "./combo-catalog-section";

export const metadata: Metadata = {
  title: "Shop",
  description: "Browse pastries, cakes, and seasonal treats for pickup.",
};

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
